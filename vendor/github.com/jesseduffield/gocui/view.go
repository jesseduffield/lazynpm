// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/go-errors/errors"

	"github.com/jesseduffield/termbox-go"
	"github.com/mattn/go-runewidth"
	"github.com/sirupsen/logrus"
)

// Constants for overlapping edges
const (
	TOP    = 1 // view is overlapping at top edge
	BOTTOM = 2 // view is overlapping at bottom edge
	LEFT   = 4 // view is overlapping at left edge
	RIGHT  = 8 // view is overlapping at right edge
)

// A View is a window. It maintains its own internal buffer and cursor
// position.
type View struct {
	name           string
	x0, y0, x1, y1 int
	ox, oy         int
	// these are the view cursor's X,Y coordinates. The view cursor determines
	// what line appears highlighted.
	cx, cy int
	// these are the write cursor's X,Y coordinates. The write cursor determines
	// where the next characters will be written to in the view.
	wcx, wcy   int
	lines      [][]cell
	readOffset int
	readCache  string

	tainted   bool       // marks if the viewBuffer must be updated
	viewLines []viewLine // internal representation of the view's buffer

	ei *escapeInterpreter // used to decode ESC sequences on Write

	// BgColor and FgColor allow to configure the background and foreground
	// colors of the View.
	BgColor, FgColor Attribute

	// SelBgColor and SelFgColor are used to configure the background and
	// foreground colors of the selected line, when it is highlighted.
	SelBgColor, SelFgColor Attribute

	// If Editable is true, keystrokes will be added to the view's internal
	// buffer at the cursor position.
	Editable bool

	// Editor allows to define the editor that manages the edition mode,
	// including keybindings or cursor behaviour. DefaultEditor is used by
	// default.
	Editor Editor

	// Overwrite enables or disables the overwrite mode of the view.
	Overwrite bool

	// If Highlight is true, Sel{Bg,Fg}Colors will be used
	// for the line under the cursor position.
	Highlight bool

	// If Frame is true, a border will be drawn around the view.
	Frame bool

	// If Wrap is true, the content that is written to this View is
	// automatically wrapped when it is longer than its width. If true the
	// view's x-origin will be ignored.
	Wrap bool

	// If Autoscroll is true, the View will automatically scroll down when the
	// text overflows. If true the view's y-origin will be ignored.
	Autoscroll bool

	// If Frame is true, Title allows to configure a title for the view.
	Title string

	Tabs     []string
	TabIndex int

	// If Frame is true, Subtitle allows to configure a subtitle for the view.
	Subtitle string

	// If Mask is true, the View will display the mask instead of the real
	// content
	Mask rune

	// Overlaps describes which edges are overlapping with another view's edges
	Overlaps byte

	// If HasLoader is true, the message will be appended with a spinning loader animation
	HasLoader bool

	writeMutex sync.Mutex

	// IgnoreCarriageReturns tells us whether to ignore '\r' characters
	IgnoreCarriageReturns bool

	// ParentView is the view which catches events bubbled up from the given view if there's no matching handler
	ParentView *View

	Context string // this is for assigning keybindings to a view only in certain contexts

	searcher *searcher

	// when ContainsList is true, we show the current index and total count in the view
	ContainsList bool

	log *logrus.Entry

	// Pty tells us whether we have a pty running within the view. When Pty is set,
	// we will catch keybindings set on the view, but all other keypresses (including
	// those for which there are global keybindings) will be forwarded to the
	// underlying pty as the original byte string. When Pty is set, we will also
	// directly redraw the view when it is written to
	Pty bool

	// StdinWriter is used in conjunction with the Pty flag. When using a pty,
	// any termbox events not caught by the view will be written to this writer
	// as the original byte slice.
	StdinWriter io.Writer

	// these are for when the terminal wants to save the cursor position to restore
	// it later
	savedWcx   int
	savedWcy   int
	savedLines [][]cell // TODO: see if we need a separate savedWcx and savedWcy for dealing with code 1049
	savedOx    int
	savedOy    int

	// these are the top and bottom scroll margins
	topMargin    int
	bottomMargin int

	// IgnoreClickPosition determines whether we set the cursor upon clicking a view
	IgnoreClickPosition bool

	Visible bool
}

type searcher struct {
	searchString       string
	searchPositions    []cellPos
	currentSearchIndex int
	onSelectItem       func(int, int, int) error
}

func (v *View) SetOnSelectItem(onSelectItem func(int, int, int) error) {
	v.searcher.onSelectItem = onSelectItem
}

func (v *View) gotoNextMatch() error {
	if len(v.searcher.searchPositions) == 0 {
		return nil
	}
	if v.searcher.currentSearchIndex == len(v.searcher.searchPositions)-1 {
		v.searcher.currentSearchIndex = 0
	} else {
		v.searcher.currentSearchIndex++
	}
	return v.SelectSearchResult(v.searcher.currentSearchIndex)
}

func (v *View) gotoPreviousMatch() error {
	if len(v.searcher.searchPositions) == 0 {
		return nil
	}
	if v.searcher.currentSearchIndex == 0 {
		if len(v.searcher.searchPositions) > 0 {
			v.searcher.currentSearchIndex = len(v.searcher.searchPositions) - 1
		}
	} else {
		v.searcher.currentSearchIndex--
	}
	return v.SelectSearchResult(v.searcher.currentSearchIndex)
}

func (v *View) SelectSearchResult(index int) error {
	y := v.searcher.searchPositions[index].y
	v.FocusPoint(0, y)
	if v.searcher.onSelectItem != nil {
		return v.searcher.onSelectItem(y, index, len(v.searcher.searchPositions))
	}
	return nil
}

func (v *View) Search(str string) error {
	v.writeMutex.Lock()
	defer v.writeMutex.Unlock()

	v.searcher.search(str)
	v.updateSearchPositions()
	if len(v.searcher.searchPositions) > 0 {
		// get the first result past the current cursor
		currentIndex := 0
		adjustedY := v.oy + v.cy
		adjustedX := v.ox + v.cx
		for i, pos := range v.searcher.searchPositions {
			if pos.y > adjustedY || (pos.y == adjustedY && pos.x > adjustedX) {
				currentIndex = i
				break
			}
		}
		v.searcher.currentSearchIndex = currentIndex
		return v.SelectSearchResult(currentIndex)
	} else {
		return v.searcher.onSelectItem(-1, -1, 0)
	}
	return nil
}

func (v *View) ClearSearch() {
	v.searcher.clearSearch()
}

func (v *View) IsSearching() bool {
	return v.searcher.searchString != ""
}

func (v *View) FocusPoint(cx int, cy int) {
	lineCount := len(v.lines)
	if cy < 0 || cy > lineCount {
		return
	}
	_, height := v.Size()

	ly := height - 1
	if ly == -1 {
		ly = 0
	}

	// if line is above origin, move origin and set cursor to zero
	// if line is below origin + height, move origin and set cursor to max
	// otherwise set cursor to value - origin
	if ly > lineCount {
		v.cx = cx
		v.cy = cy
		v.oy = 0
	} else if cy < v.oy {
		v.cx = cx
		v.cy = 0
		v.oy = cy
	} else if cy > v.oy+ly {
		v.cx = cx
		v.cy = ly
		v.oy = cy - ly
	} else {
		v.cx = cx
		v.cy = cy - v.oy
	}
}

func (s *searcher) search(str string) {
	s.searchString = str
	s.searchPositions = []cellPos{}
	s.currentSearchIndex = 0
}

func (s *searcher) clearSearch() {
	s.searchString = ""
	s.searchPositions = []cellPos{}
	s.currentSearchIndex = 0
}

type cellPos struct {
	x int
	y int
}

type viewLine struct {
	linesX, linesY int // coordinates relative to v.lines
	line           []cell
}

type cell struct {
	chr              rune
	bgColor, fgColor Attribute
}

type lineType []cell

// String returns a string from a given cell slice.
func (l lineType) String() string {
	str := ""
	for _, c := range l {
		str += string(c.chr)
	}
	return str
}

// newView returns a new View object.
func newView(name string, x0, y0, x1, y1 int, mode OutputMode, log *logrus.Entry) *View {
	v := &View{
		searcher:     &searcher{},
		name:         name,
		x0:           x0,
		y0:           y0,
		x1:           x1,
		y1:           y1,
		Frame:        true,
		Editor:       DefaultEditor,
		tainted:      true,
		ei:           newEscapeInterpreter(mode),
		log:          log,
		topMargin:    1,
		bottomMargin: y1 - y0, // TODO: this might be off by one
		Visible:      true,
	}
	return v
}

// Dimensions returns the dimensions of the View
func (v *View) Dimensions() (int, int, int, int) {
	return v.x0, v.y0, v.x1, v.y1
}

// Size returns the number of visible columns and rows in the View.
func (v *View) Size() (x, y int) {
	return v.x1 - v.x0 - 1, v.y1 - v.y0 - 1
}

// Name returns the name of the view.
func (v *View) Name() string {
	return v.name
}

// setRune sets a rune at the given point relative to the view. It applies the
// specified colors, taking into account if the cell must be highlighted. Also,
// it checks if the position is valid.
func (v *View) setRune(x, y int, ch rune, fgColor, bgColor Attribute) error {
	maxX, maxY := v.Size()
	if x < 0 || x >= maxX || y < 0 || y >= maxY {
		return errors.New("invalid point")
	}
	var (
		ry, rcy int
		err     error
	)
	if v.Highlight {
		_, ry, err = v.realPosition(x, y)
		if err != nil {
			return err
		}
		_, rcy, err = v.realPosition(v.cx, v.cy)
		if err != nil {
			return err
		}
	}

	if v.Mask != 0 {
		fgColor = v.FgColor
		bgColor = v.BgColor
		ch = v.Mask
	} else if v.Highlight && ry == rcy {
		fgColor = fgColor | AttrBold
	}

	termbox.SetCell(v.x0+x+1, v.y0+y+1, ch,
		termbox.Attribute(fgColor), termbox.Attribute(bgColor))

	return nil
}

// SetCursor sets the cursor position of the view at the given point,
// relative to the view. It checks if the position is valid.
func (v *View) SetCursor(x, y int) error {
	if x < 0 || y < 0 {
		return nil
	}
	v.cx = x
	v.cy = y
	return nil
}

// Cursor returns the cursor position of the view.
func (v *View) Cursor() (x, y int) {
	return v.cx, v.cy
}

// SetOrigin sets the origin position of the view's internal buffer,
// so the buffer starts to be printed from this point, which means that
// it is linked with the origin point of view. It can be used to
// implement Horizontal and Vertical scrolling with just incrementing
// or decrementing ox and oy.
func (v *View) SetOrigin(x, y int) error {
	v.ox = x
	v.oy = y
	return nil
}

// Origin returns the origin position of the view.
func (v *View) Origin() (x, y int) {
	return v.ox, v.oy
}

func (v *View) padCellsForNewCy() {
	if v.wcx >= len(v.lines[v.wcy]) {
		v.lines[v.wcy] = append(v.lines[v.wcy], make([]cell, v.wcx-len(v.lines[v.wcy]))...)
	}
}

func (v *View) moveCursorHorizontally(n int) {
	if n > 0 {
		v.moveCursorRight(n)
		return
	}
	v.moveCursorLeft(-n)
}

func (v *View) moveCursorVertically(n int) {
	if n > 0 {
		v.moveCursorDown(n)
		return
	}
	v.moveCursorUp(-n)
}

func (v *View) moveCursorRight(n int) {
	for i := 0; i < n; i++ {
		if v.wcx == len(v.lines[v.wcy]) {
			v.lines[v.wcy] = append(v.lines[v.wcy], cell{})
		}
		v.wcx++
	}
}

func (v *View) moveCursorLeft(n int) {
	if v.wcx-n <= 0 {
		v.wcx = 0
	} else {
		v.wcx -= n
	}
}

func (v *View) moveCursorDown(n int) {
	for i := 0; i < n; i++ {
		if v.wcy == len(v.lines)-1 {
			v.lines = append(v.lines, nil)
		}
		v.wcy++
	}
	v.padCellsForNewCy()
}

func (v *View) moveCursorUp(n int) {
	if v.wcy-n <= 0 {
		v.wcy = 0
	} else {
		v.wcy -= n
	}
	v.padCellsForNewCy()
}

func (v *View) moveCursorToPosition(x int, y int) {
	v.moveCursorVertically(y - v.wcy)
	v.moveCursorHorizontally(x - v.wcx)
}

func quoteRunes(runes []rune) string {
	str := ""
	for _, r := range runes {
		str += strconv.QuoteRune(r)
	}
	return str
}

func insertLine(lines [][]cell, line []cell, index int) [][]cell {
	// TODO: handle padding
	lines = append(lines, nil)
	copy(lines[index+1:], lines[index:])
	lines[index] = line
	return lines
}

func deleteLine(lines [][]cell, index int) [][]cell {
	if len(lines) < index {
		return lines
	}
	if index < len(lines)-1 {
		copy(lines[index:], lines[index+1:])
	}
	lines[len(lines)-1] = nil
	lines = lines[:len(lines)-1]
	return lines
}

// Write appends a byte slice into the view's internal buffer. Because
// View implements the io.Writer interface, it can be passed as parameter
// of functions like fmt.Fprintf, fmt.Fprintln, io.Copy, etc. Clear must
// be called to clear the view's buffer.
func (v *View) Write(p []byte) (n int, err error) {
	v.tainted = true
	v.writeMutex.Lock()
	defer v.writeMutex.Unlock()

	if len(v.lines) == 0 {
		v.lines = [][]cell{[]cell{cell{}}}
	}

	sanityCheck := func() {
		if v.lines == nil {
			panic("v.lines == nil")
		}
		if v.wcx > len(v.lines[v.wcy]) {
			panic("cx too big")
		}
	}

	runes := bytes.Runes(p)
	for _, ch := range runes {
		switch ch {
		case '\n':
			v.wcx = 0
			v.wcy += 1
			if v.wcy == len(v.lines) {
				v.lines = append(v.lines, []cell{})
			} else if v.wcy == v.bottomMargin {
				v.lines = deleteLine(v.lines, v.topMargin-1)
				v.lines = insertLine(v.lines, []cell{}, v.bottomMargin-1)
			}
			sanityCheck()
		case '\r':
			if v.IgnoreCarriageReturns {
				continue
			}
			v.wcx = 0

			v.ei.reset()

			sanityCheck()
		default:

			cells := v.parseInput(ch)
			if v.ei.instruction.kind != NONE {
				switch v.ei.instruction.kind {
				case CURSOR_UP:
					toMoveUp := v.ei.instruction.param1
					v.moveCursorUp(toMoveUp)
					sanityCheck()
				case CURSOR_DOWN:
					toMoveDown := v.ei.instruction.param1
					v.moveCursorDown(toMoveDown)
					sanityCheck()

				case CURSOR_LEFT:
					toMoveLeft := v.ei.instruction.param1
					v.moveCursorLeft(toMoveLeft)
					sanityCheck()
				case CURSOR_RIGHT:
					toMoveRight := v.ei.instruction.param1
					v.moveCursorRight(toMoveRight)
					sanityCheck()
				case CURSOR_MOVE:
					y := v.ei.instruction.param1
					x := v.ei.instruction.param2
					// these params are 1-indexed so we need to decrement them (0 is left alone)
					if x > 0 {
						x--
					}
					if y > 0 {
						y--
					}

					v.moveCursorToPosition(x, y)
					sanityCheck()
				case ERASE_IN_LINE:
					code := v.ei.instruction.param1
					// need to check if we should delete the character at the cursor as well. I will assume that we don't (cos it's easier code-wise)
					switch code {
					case 0:
						v.lines[v.wcy] = v.lines[v.wcy][0:v.wcx]
						sanityCheck()
					case 1:
						v.lines[v.wcy] = append(make([]cell, v.wcx), v.lines[v.wcy][v.wcx:]...)
						sanityCheck()
					case 2:
						// need to clear line but retain cursor x position
						// so we'll pad everything out to the left
						v.lines[v.wcy] = make([]cell, v.wcx+1)
						sanityCheck()
					}

				case CLEAR_SCREEN:
					code := v.ei.instruction.param1

					switch code {
					case 0:
						// erase from current position to end of screen, left to right and up to down
						v.lines[v.wcy] = v.lines[v.wcy][0:v.wcx]
						v.lines[v.wcy] = append(v.lines[v.wcy], cell{})
						if len(v.lines)-1 > v.wcy {
							v.lines = v.lines[:v.wcy+1]
						}
						sanityCheck()
					case 1:
						// TODO: test
						if v.wcy > 0 {
							v.lines = append(make([][]cell, v.wcy), v.lines[v.wcy:]...)
						}
						v.lines[v.wcy] = append(make([]cell, v.wcx), v.lines[v.wcy][v.wcx:]...)
						sanityCheck()
					case 2:
						v.lines = make([][]cell, 1)
						// TODO: apparently the cursor isn't actually supposed to move here. We'll need to pad this out.
						v.wcx = 0
						v.wcy = 0
						v.ox = 0
						v.oy = 0
						sanityCheck()
					}

				case INSERT_CHARACTER:
					toInsert := v.ei.instruction.param1
					if toInsert == 0 {
						toInsert = 1
					}
					v.lines[v.wcy] = append(v.lines[v.wcy][:v.wcx], append(make([]cell, toInsert), v.lines[v.wcy][v.wcx:]...)...)
				case DELETE:
					toDelete := v.ei.instruction.param1
					if toDelete == 0 {
						toDelete = 1
					}
					if v.wcx+toDelete > len(v.lines[v.wcy]) {
						toDelete = len(v.lines[v.wcy]) - v.wcx
					}
					v.lines[v.wcy] = append(v.lines[v.wcy][:v.wcx], v.lines[v.wcy][v.wcx+toDelete:]...)
				case SAVE_CURSOR_POSITION:
					v.savedWcx = v.wcx
					v.savedWcy = v.wcy
				case RESTORE_CURSOR_POSITION:
					v.moveCursorToPosition(v.savedWcx, v.savedWcy)
					sanityCheck()
				case WRITE:
					v.StdinWriter.Write([]byte(string(v.ei.instruction.toWrite)))
				case SWITCH_TO_ALTERNATE_SCREEN:
					v.savedLines = v.lines
					v.lines = [][]cell{[]cell{cell{}}}
					v.savedWcx = v.wcx
					v.savedWcy = v.wcy
					v.savedOx = v.ox
					v.savedOy = v.oy
					v.wcy = 0
					v.wcx = 0
					v.oy = 0
					v.ox = 0
					v.Autoscroll = false
					sanityCheck()
				case SWITCH_BACK_FROM_ALTERNATE_SCREEN:
					v.lines = v.savedLines
					v.topMargin = 1
					v.bottomMargin = v.y1 - v.y0
					v.wcx = v.savedWcx
					v.wcy = v.savedWcy
					v.ox = v.savedOx
					v.oy = v.savedOy
					v.Autoscroll = true
				case SET_SCROLL_MARGINS:
					v.topMargin = v.ei.instruction.param1
					v.bottomMargin = v.ei.instruction.param2
				case INSERT_LINES:
					if v.wcy+1 < v.topMargin || v.wcy+1 > v.bottomMargin {
						continue
					}
					for i := 0; i < v.ei.instruction.param1; i++ {
						if len(v.lines) >= v.bottomMargin {
							v.lines = deleteLine(v.lines, v.bottomMargin-1)
						}

						v.lines = insertLine(v.lines, []cell{}, v.wcy)
					}
					sanityCheck()
				case DELETE_LINES:

					if v.wcy+1 < v.topMargin || v.wcy+1 > v.bottomMargin {
						continue
					}
					for i := 0; i < v.ei.instruction.param1; i++ {
						v.lines = insertLine(v.lines, []cell{}, v.bottomMargin-1)
						v.lines = deleteLine(v.lines, v.wcy)
					}
					sanityCheck()
				default:
					panic("instruction not understood")
				}
				v.ei.instructionRead()
				continue
			}
			if cells == nil {
				continue
			}

			for _, c := range cells {
				if c.chr == 7 {
					// bell: can't do anything
					continue
				}
				if c.chr == '\b' {
					if v.wcx > 0 {
						v.wcx--
					}
					continue
				}
				if v.wcx == len(v.lines[v.wcy]) {
					v.lines[v.wcy] = append(v.lines[v.wcy], c)
				} else if v.wcx < len(v.lines[v.wcy]) {
					v.lines[v.wcy][v.wcx] = c
				} else {
					// TODO: decide whether this matters
					panic(v.name + ": above length for some reason")
				}

				v.wcx++

				if v.Wrap {
					width, _ := v.Size()
					if v.wcx == width {
						v.moveCursorDown(1)
						v.wcx = 0
					}
				}
			}
			sanityCheck()
		}
	}

	return len(p), nil
}

// parseInput parses char by char the input written to the View. It returns nil
// while processing ESC sequences. Otherwise, it returns a cell slice that
// contains the processed data.
func (v *View) parseInput(ch rune) []cell {
	cells := []cell{}

	isEscape, err := v.ei.parseOne(ch)
	if err != nil {
		// there is an error parsing an escape sequence, ouput all the escape characters so far as a string
		for _, r := range v.ei.runes() {
			c := cell{
				fgColor: v.FgColor,
				bgColor: v.BgColor,
				chr:     r,
			}
			cells = append(cells, c)
		}
		v.ei.reset()
	} else {
		if isEscape {
			return nil
		}
		repeatCount := 1
		if ch == '\t' {
			ch = ' '
			repeatCount = 4
		}
		for i := 0; i < repeatCount; i++ {
			c := cell{
				fgColor: v.ei.curFgColor,
				bgColor: v.ei.curBgColor,
				chr:     ch,
			}
			cells = append(cells, c)
		}
	}

	return cells
}

// Read reads data into p. It returns the number of bytes read into p.
// At EOF, err will be io.EOF. Calling Read() after Rewind() makes the
// cache to be refreshed with the contents of the view.
func (v *View) Read(p []byte) (n int, err error) {
	if v.readOffset == 0 {
		v.readCache = v.Buffer()
	}
	if v.readOffset < len(v.readCache) {
		n = copy(p, v.readCache[v.readOffset:])
		v.readOffset += n
	} else {
		err = io.EOF
	}
	return
}

// Rewind sets the offset for the next Read to 0, which also refresh the
// read cache.
func (v *View) Rewind() {
	v.readOffset = 0
}

func containsUpcaseChar(str string) bool {
	for _, ch := range str {
		if unicode.IsUpper(ch) {
			return true
		}
	}
	return false
}

func (v *View) updateSearchPositions() {
	if v.searcher.searchString != "" {
		var normalizeRune func(r rune) rune
		var normalizedSearchStr string
		// if we have any uppercase characters we'll do a case-sensitive search
		if containsUpcaseChar(v.searcher.searchString) {
			normalizedSearchStr = v.searcher.searchString
			normalizeRune = func(r rune) rune { return r }
		} else {
			normalizedSearchStr = strings.ToLower(v.searcher.searchString)
			normalizeRune = unicode.ToLower
		}

		v.searcher.searchPositions = []cellPos{}
		for y, line := range v.lines {
		lineLoop:
			for x, _ := range line {
				if normalizeRune(line[x].chr) == rune(normalizedSearchStr[0]) {
					for offset := 1; offset < len(normalizedSearchStr); offset++ {
						if len(line)-1 < x+offset {
							continue lineLoop
						}
						if normalizeRune(line[x+offset].chr) != rune(normalizedSearchStr[offset]) {
							continue lineLoop
						}
					}
					v.searcher.searchPositions = append(v.searcher.searchPositions, cellPos{x: x, y: y})
				}
			}
		}
	}

}

// draw re-draws the view's contents. It returns an array of wrap counts, that is,
// when wrapping is turned on, an array where for each index and value i, v,
// i is the position of a line in the buffer, and v is the number times the line wrapped
func (v *View) draw() ([]int, error) {
	v.writeMutex.Lock()
	defer v.writeMutex.Unlock()

	v.updateSearchPositions()
	maxX, maxY := v.Size()

	if v.Wrap {
		if maxX == 0 {
			return nil, errors.New("X size of the view cannot be 0")
		}
		v.ox = 0
	}

	wrapCounts := make([]int, len(v.lines))

	if v.tainted {
		v.viewLines = nil
		lines := v.lines
		if v.HasLoader {
			lines = v.loaderLines()
		}
		for i, line := range lines {
			wrap := 0
			if v.Wrap {
				// TODO: see if we should consider v.Frame here
				wrap = maxX + 1
			}

			ls := lineWrap(line, wrap)
			wrapCounts[i] = len(ls) - 1
			for j := range ls {
				vline := viewLine{linesX: j, linesY: i, line: ls[j]}
				v.viewLines = append(v.viewLines, vline)
			}
		}
		if !v.HasLoader {
			v.tainted = false
		}
	}

	if v.Autoscroll && len(v.viewLines) > maxY {
		v.oy = len(v.viewLines) - maxY
	}
	y := 0
	for i, vline := range v.viewLines {
		if i < v.oy {
			continue
		}
		if y >= maxY {
			break
		}
		x := 0
		for j, c := range vline.line {
			if j < v.ox {
				continue
			}
			if x >= maxX {
				break
			}

			fgColor := c.fgColor
			if fgColor == ColorDefault {
				fgColor = v.FgColor
			}
			bgColor := c.bgColor
			if bgColor == ColorDefault {
				bgColor = v.BgColor
			}
			if matched, selected := v.isPatternMatchedRune(x, y); matched {
				if selected {
					bgColor = ColorCyan
				} else {
					bgColor = ColorYellow
				}
			}

			if err := v.setRune(x, y, c.chr, fgColor, bgColor); err != nil {
				return nil, err
			}
			x += runewidth.RuneWidth(c.chr)
		}
		y++
	}
	return wrapCounts, nil
}

func (v *View) isPatternMatchedRune(x, y int) (bool, bool) {
	searchStringLength := len(v.searcher.searchString)
	for i, pos := range v.searcher.searchPositions {
		adjustedY := y + v.oy
		adjustedX := x + v.ox
		if adjustedY == pos.y && adjustedX >= pos.x && adjustedX < pos.x+searchStringLength {
			return true, i == v.searcher.currentSearchIndex
		}
	}
	return false, false
}

// realPosition returns the position in the internal buffer corresponding to the
// point (x, y) of the view.
func (v *View) realPosition(vx, vy int) (x, y int, err error) {
	vx = v.ox + vx
	vy = v.oy + vy

	if vx < 0 || vy < 0 {
		return 0, 0, errors.New("invalid point")
	}

	if len(v.viewLines) == 0 {
		return vx, vy, nil
	}

	if vy < len(v.viewLines) {
		vline := v.viewLines[vy]
		x = vline.linesX + vx
		y = vline.linesY
	} else {
		vline := v.viewLines[len(v.viewLines)-1]
		x = vx
		y = vline.linesY + vy - len(v.viewLines) + 1
	}

	return x, y, nil
}

// Clear empties the view's internal buffer.
func (v *View) Clear() {
	v.writeMutex.Lock()
	defer v.writeMutex.Unlock()

	v.tainted = true
	v.ei.reset()

	v.lines = nil
	v.wcy = 0
	v.wcx = 0
	v.viewLines = nil
	v.readOffset = 0
	v.clearRunes()
}

// clearRunes erases all the cells in the view.
func (v *View) clearRunes() {
	maxX, maxY := v.Size()
	for x := 0; x < maxX; x++ {
		for y := 0; y < maxY; y++ {
			termbox.SetCell(v.x0+x+1, v.y0+y+1, ' ',
				termbox.Attribute(v.FgColor), termbox.Attribute(v.BgColor))
		}
	}
}

// BufferLines returns the lines in the view's internal
// buffer.
func (v *View) BufferLines() []string {
	v.writeMutex.Lock()
	defer v.writeMutex.Unlock()
	lines := make([]string, len(v.lines))
	for i, l := range v.lines {
		str := lineType(l).String()
		str = strings.Replace(str, "\x00", " ", -1)
		lines[i] = str
	}
	return lines
}

// Buffer returns a string with the contents of the view's internal
// buffer.
func (v *View) Buffer() string {
	return linesToString(v.lines)
}

// ViewBufferLines returns the lines in the view's internal
// buffer that is shown to the user.
func (v *View) ViewBufferLines() []string {
	v.writeMutex.Lock()
	defer v.writeMutex.Unlock()
	lines := make([]string, len(v.viewLines))
	for i, l := range v.viewLines {
		str := lineType(l.line).String()
		str = strings.Replace(str, "\x00", " ", -1)
		lines[i] = str
	}
	return lines
}

// LinesHeight is the count of view lines (i.e. lines excluding wrapping)
func (v *View) LinesHeight() int {
	return len(v.lines)
}

// ViewLinesHeight is the count of view lines (i.e. lines including wrapping)
func (v *View) ViewLinesHeight() int {
	return len(v.viewLines)
}

// ViewBuffer returns a string with the contents of the view's buffer that is
// shown to the user.
func (v *View) ViewBuffer() string {
	lines := make([][]cell, len(v.viewLines))
	for i := range v.viewLines {
		lines[i] = v.viewLines[i].line
	}

	return linesToString(lines)
}

// Line returns a string with the line of the view's internal buffer
// at the position corresponding to the point (x, y).
func (v *View) Line(y int) (string, error) {
	_, y, err := v.realPosition(0, y)
	if err != nil {
		return "", err
	}

	if y < 0 || y >= len(v.lines) {
		return "", errors.New("invalid point")
	}

	return lineType(v.lines[y]).String(), nil
}

// Word returns a string with the word of the view's internal buffer
// at the position corresponding to the point (x, y).
func (v *View) Word(x, y int) (string, error) {
	x, y, err := v.realPosition(x, y)
	if err != nil {
		return "", err
	}

	if x < 0 || y < 0 || y >= len(v.lines) || x >= len(v.lines[y]) {
		return "", errors.New("invalid point")
	}

	str := lineType(v.lines[y]).String()

	nl := strings.LastIndexFunc(str[:x], indexFunc)
	if nl == -1 {
		nl = 0
	} else {
		nl = nl + 1
	}
	nr := strings.IndexFunc(str[x:], indexFunc)
	if nr == -1 {
		nr = len(str)
	} else {
		nr = nr + x
	}
	return string(str[nl:nr]), nil
}

// indexFunc allows to split lines by words taking into account spaces
// and 0.
func indexFunc(r rune) bool {
	return r == ' ' || r == 0
}

func lineWidth(line []cell) (n int) {
	for i := range line {
		n += runewidth.RuneWidth(line[i].chr)
	}

	return
}

func lineWrap(line []cell, columns int) [][]cell {
	if columns == 0 {
		return [][]cell{line}
	}

	var n int
	var offset int
	lines := make([][]cell, 0, 1)
	for i := range line {
		rw := runewidth.RuneWidth(line[i].chr)
		n += rw
		if n > columns {
			n = rw
			lines = append(lines, line[offset:i])
			offset = i
		}
	}

	lines = append(lines, line[offset:])
	return lines
}

func linesToString(lines [][]cell) string {
	str := make([]string, len(lines))
	for i := range lines {
		rns := make([]rune, 0, len(lines[i]))
		line := lineType(lines[i]).String()
		for _, c := range line {
			if c != '\x00' {
				rns = append(rns, c)
			}
		}
		str[i] = string(rns)
	}

	return strings.Join(str, "\n")
}

func (v *View) loaderLines() [][]cell {
	duplicate := make([][]cell, len(v.lines))
	for i := range v.lines {
		if i < len(v.lines)-1 {
			duplicate[i] = make([]cell, len(v.lines[i]))
			copy(duplicate[i], v.lines[i])
		} else {
			duplicate[i] = make([]cell, len(v.lines[i])+2)
			copy(duplicate[i], v.lines[i])
			duplicate[i][len(duplicate[i])-2] = cell{chr: ' '}
			duplicate[i][len(duplicate[i])-1] = Loader()
		}
	}

	return duplicate
}

func Loader() cell {
	characters := "|/-\\"
	now := time.Now()
	nanos := now.UnixNano()
	index := nanos / 50000000 % int64(len(characters))
	str := characters[index : index+1]
	chr := []rune(str)[0]
	return cell{
		chr: chr,
	}
}

// IsTainted tells us if the view is tainted
func (v *View) IsTainted() bool {
	return v.tainted
}

// GetClickedTabIndex tells us which tab was clicked
func (v *View) GetClickedTabIndex(x int) int {
	if len(v.Tabs) <= 1 {
		return 0
	}

	charIndex := 0
	for i, tab := range v.Tabs {
		charIndex += len(tab + " - ")
		if x < charIndex {
			return i
		}
	}

	return 0
}

func (v *View) SelectedLineIdx() int {
	_, seletedLineIdx := v.SelectedPoint()
	return seletedLineIdx
}

func (v *View) SelectedPoint() (int, int) {
	// TODO: I feel like we should have to involve the origin here but it seems
	// like the meaning of the cursor has changed from being the position relative to the origin
	// vs the absolute position in terms of the view's content. Will need to check ramifications of this
	cx, cy := v.Cursor()
	return cx, cy
}
