// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocui

import (
	"strconv"
	"strings"
	"sync"

	"github.com/go-errors/errors"
)

type escapeInterpreter struct {
	state                  escapeState
	curch                  rune
	csiParam               []string
	oscParam               []rune
	curFgColor, curBgColor Attribute
	mode                   OutputMode
	mutex                  sync.Mutex
	instruction            instruction
}

/// new algorithm: if we hit an escape key start a sequence
// if we then hit a terminator key, process the sequence

const (
	NONE = 1 << iota
	CURSOR_UP
	CURSOR_DOWN
	CURSOR_LEFT
	CURSOR_RIGHT
	CURSOR_MOVE
	CLEAR_SCREEN
	ERASE_IN_LINE
	INSERT_CHARACTER
	DELETE
	SAVE_CURSOR_POSITION
	RESTORE_CURSOR_POSITION
	WRITE
	SWITCH_TO_ALTERNATE_SCREEN
	SWITCH_BACK_FROM_ALTERNATE_SCREEN
	SET_SCROLL_MARGINS
	INSERT_LINES
	DELETE_LINES
)

type instruction struct {
	kind    int
	param1  int
	param2  int
	toWrite []rune
}

type escapeState int

const (
	stateNone escapeState = iota
	stateEscape
	stateCSI
	stateParams
	stateCharacterSet
	stateOSC
)

var directionMap = map[rune]int{
	'A': CURSOR_UP,
	'B': CURSOR_DOWN,
	'C': CURSOR_RIGHT,
	'D': CURSOR_LEFT,
}

var (
	errNotCSI        = errors.New("Not a CSI escape sequence")
	errCSIParseError = errors.New("CSI escape sequence parsing error")
	errCSITooLong    = errors.New("CSI escape sequence is too long")
)

// runes in case of error will output the non-parsed runes as a string.
func (ei *escapeInterpreter) runes() []rune {
	ei.mutex.Lock()
	defer ei.mutex.Unlock()

	switch ei.state {
	case stateNone:
		return []rune{0x1b}
	case stateEscape:
		return []rune{0x1b, ei.curch}
	case stateCSI:
		return []rune{0x1b, '[', ei.curch}
	case stateParams:
		ret := []rune{0x1b, '['}
		for i, s := range ei.csiParam {
			ret = append(ret, []rune(s)...)
			if i < len(ei.csiParam)-1 {
				ret = append(ret, ';')
			}
		}
		return append(ret, ei.curch)
	}
	return nil
}

// newEscapeInterpreter returns an escapeInterpreter that will be able to parse
// terminal escape sequences.
func newEscapeInterpreter(mode OutputMode) *escapeInterpreter {
	ei := &escapeInterpreter{
		state:       stateNone,
		curFgColor:  ColorDefault,
		curBgColor:  ColorDefault,
		mode:        mode,
		instruction: instruction{kind: NONE},
	}
	return ei
}

// reset sets the escapeInterpreter in initial state.
func (ei *escapeInterpreter) reset() {
	ei.mutex.Lock()
	defer ei.mutex.Unlock()

	ei.state = stateNone
	ei.curFgColor = ColorDefault
	ei.curBgColor = ColorDefault
	ei.csiParam = nil
	ei.oscParam = nil
}

func (ei *escapeInterpreter) instructionRead() {
	ei.instruction.kind = NONE
}

func (ei *escapeInterpreter) parseOSCParams() {
	str := string(ei.oscParam)
	params := strings.Split(str, ";")
	switch params[0] {
	case "0", "2":
		// ignoring attempts at setting the title
		return
	case "11":
		// ignoring for now
		// ei.instruction.kind = WRITE
		// ei.instruction.toWrite = []rune{0x1b, ']', '1', '1', ';', 'r', 'g', 'b', ':', '0', '0', '/', '0', '0', '/', '0', '0', 0x1b, '\\'}
	default:
		panic(str)
	}
}

// parseOne parses a rune. If isEscape is true, it means that the rune is part
// of an escape sequence, and as such should not be printed verbatim. Otherwise,
// it's not an escape sequence.
func (ei *escapeInterpreter) parseOne(ch rune) (isEscape bool, err error) {
	ei.mutex.Lock()
	defer ei.mutex.Unlock()

	// Sanity checks
	if len(ei.csiParam) > 20 {
		return false, errCSITooLong
	}
	if len(ei.csiParam) > 0 && len(ei.csiParam[len(ei.csiParam)-1]) > 255 {
		return false, errCSITooLong
	}

	ei.curch = ch

	switch ei.state {
	case stateNone:
		if ch == 0x1b {
			ei.state = stateEscape
			return true, nil
		}
		return false, nil
	case stateEscape:
		switch ch {
		case '[':
			ei.state = stateCSI
			return true, nil
		case ']':
			ei.state = stateOSC
			return true, nil
		case '(':
			ei.state = stateCharacterSet
			return true, nil
		case '=':
			// set Keypad Numeric Mode
			// not implemented
			ei.state = stateNone
			return true, nil
		case '>':
			// set Keypad Application Mode
			// not implemented
			ei.state = stateNone
			return true, nil
		default:
			return false, errNotCSI
		}
	case stateOSC:
		// either we have a bell in which case we parse what we've gotten so far
		// or we append to our params
		if ch == '\a' {
			ei.parseOSCParams()
			ei.state = stateNone
			return true, nil
		} else {
			ei.oscParam = append(ei.oscParam, ch)
			return true, nil
		}
	case stateCharacterSet:
		// doing nothing for now.
		// see 'ESC ( C ' in https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-Functions-using-CSI-_-ordered-by-the-final-character_s_
		ei.state = stateNone
		return true, nil
	case stateCSI:
		switch {
		case ch >= '0' && ch <= '9', ch == '?', ch == '>', ch == ';':
			ei.csiParam = append(ei.csiParam, "")
		case ch == 'A', ch == 'B', ch == 'C', ch == 'D':
			ei.csiParam = append(ei.csiParam, "1")
		case ch == 'H', ch == 'f', ch == 'm', ch == 'K', ch == 'J':
			ei.csiParam = append(ei.csiParam, "0")
		case ch == 's':
			ei.instruction.kind = SAVE_CURSOR_POSITION
			ei.state = stateNone
			ei.csiParam = nil
			return true, nil
		case ch == 'u':
			ei.instruction.kind = RESTORE_CURSOR_POSITION
			ei.state = stateNone
			ei.csiParam = nil
			return true, nil
		case ch == 'L':
			ei.instruction.param1 = 1
			ei.instruction.kind = INSERT_LINES

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil
		case ch == 'M':
			ei.instruction.param1 = 1
			ei.instruction.kind = DELETE_LINES

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil
		default:
			return false, errCSIParseError
		}
		ei.state = stateParams
		fallthrough
	case stateParams:
		switch {
		case ch >= '0' && ch <= '9', ch == '?', ch == '>':
			ei.csiParam[len(ei.csiParam)-1] += string(ch)
			return true, nil

		case ch == ';':
			ei.csiParam = append(ei.csiParam, "")
			return true, nil

		case ch == 'm':
			var err error
			switch ei.mode {
			case OutputNormal:
				err = ei.outputNormal()
			case Output256:
				err = ei.output256()
			}
			if err != nil {
				return false, errCSIParseError
			}

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil

		case ch == 'A', ch == 'B', ch == 'C', ch == 'D':
			p, err := strconv.Atoi(ei.csiParam[0])
			if err != nil {
				return false, errCSIParseError
			}

			ei.instruction.kind = directionMap[ch]
			ei.instruction.param1 = p

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil

		case ch == 'J':
			ei.instruction.kind = CLEAR_SCREEN

			ei.instruction.param1, err = strconv.Atoi(ei.csiParam[0])
			if err != nil {
				return false, errCSIParseError
			}

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil

		case ch == 'S':
			panic("got an S")

		case ch == 'K':
			p, err := strconv.Atoi(ei.csiParam[0])
			if err != nil {
				return false, errCSIParseError
			}
			ei.instruction.kind = ERASE_IN_LINE
			ei.instruction.param1 = p

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil
		case ch == 'r':
			p, err := strconv.Atoi(ei.csiParam[0])
			if err != nil {
				return false, errCSIParseError
			}
			p2, err := strconv.Atoi(ei.csiParam[1])
			if err != nil {
				return false, errCSIParseError
			}
			ei.instruction.kind = SET_SCROLL_MARGINS
			ei.instruction.param1 = p
			ei.instruction.param2 = p2

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil
		case ch == 'L':
			ei.instruction.param1 = 1
			if len(ei.csiParam) > 0 {
				ei.instruction.param1, err = strconv.Atoi(ei.csiParam[0])
				if err != nil {
					return false, errCSIParseError
				}
			}

			ei.instruction.kind = INSERT_LINES

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil
		case ch == 'M':
			ei.instruction.param1 = 1
			if len(ei.csiParam) > 0 {
				ei.instruction.param1, err = strconv.Atoi(ei.csiParam[0])
				if err != nil {
					return false, errCSIParseError
				}
			}

			ei.instruction.kind = DELETE_LINES

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil

		case ch == 'h':
			// see https://vt100.net/docs/vt100-ug/chapter3.html
			switch ei.csiParam[0] {
			// case "?25":
			// 	// wants us to show the cursor but we never hide it in the first place
			// case "?2004":
			// 	// it's trying to set bracketed paste mode. Not sure what to do here
			case "?1049":
				ei.instruction.kind = SWITCH_TO_ALTERNATE_SCREEN
			// case "?1":
			// 	// setting cursor key to application. unimplemented
			// case "?12":
			// 	// see https://www.inwap.com/pdp10/ansicode.txt
			// 	// [12h = SRM - Send Receive Mode, transmit without local echo
			// 	// unimplemented
			// case "?1000":
			// 	// unimplemented
			// case "?1015":
			// 	// unimplemented
			// case "?1006":
			// 	// unimplemented
			// case "?1002":
			// 	// unimplemented
			default:
				ei.state = stateNone
				ei.csiParam = nil
				return true, nil

				return false, errCSIParseError
			}
			ei.state = stateNone
			ei.csiParam = nil
			return true, nil

		case ch == 'l':
			// see https://vt100.net/docs/vt100-ug/chapter3.html
			switch ei.csiParam[0] {
			// case "?25":
			// 	// wants us to show the cursor but we never hide it in the first place
			// case "?2004":
			// 	// it's trying to set bracketed paste mode. Not sure what to do here
			case "?1049":
				ei.instruction.kind = SWITCH_BACK_FROM_ALTERNATE_SCREEN
				// switch to alternate screen. unimplemented
			// case "?1":
			// 	// setting cursor key to application. unimplemented
			// case "?12":
			// 	// see https://www.inwap.com/pdp10/ansicode.txt
			// 	// [12h = SRM - Send Receive Mode, transmit without local echo
			// 	// unimplemented
			// case "?1000":
			// 	// unimplemented
			// case "?1015":
			// 	// unimplemented
			// case "?1006":
			// 	// unimplemented
			// case "?1002":
			// 	// unimplemented
			default:
				ei.state = stateNone
				ei.csiParam = nil
				return true, nil

				return false, errCSIParseError
			}
			ei.state = stateNone
			ei.csiParam = nil
			return true, nil

		case ch == 'c':
			if ei.csiParam[0] != ">" {
				return false, errCSIParseError
			}
			// I don't think we need to care about this. See https://stackoverflow.com/questions/29939026/what-is-the-ansi-escape-code-sequence-escc
			ei.state = stateNone
			ei.csiParam = nil
			return true, nil
		case ch == '@':
			p, err := strconv.Atoi(ei.csiParam[0])
			if err != nil {
				return false, errCSIParseError
			}

			if p > 1 {
				panic("don't yet support inserting more than one character")
			}

			ei.instruction.kind = INSERT_CHARACTER
			ei.instruction.param1 = p

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil

		case ch == 'P':
			p, err := strconv.Atoi(ei.csiParam[0])
			if err != nil {
				return false, errCSIParseError
			}

			ei.instruction.kind = DELETE
			ei.instruction.param1 = p

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil

		case ch == 'H', ch == 'f':
			y := 0
			if len(ei.csiParam) > 0 {
				y, err = strconv.Atoi(ei.csiParam[0])
				if err != nil {
					return false, errCSIParseError
				}
			}

			x := 0
			if len(ei.csiParam) > 1 {
				x, err = strconv.Atoi(ei.csiParam[1])
				if err != nil {
					return false, errCSIParseError
				}
			}

			ei.instruction.kind = CURSOR_MOVE
			ei.instruction.param1 = y
			ei.instruction.param2 = x

			ei.state = stateNone
			ei.csiParam = nil
			return true, nil

		default:
			return false, errCSIParseError
		}
	}
	return false, nil
}

// outputNormal provides 8 different colors:
//   black, red, green, yellow, blue, magenta, cyan, white
func (ei *escapeInterpreter) outputNormal() error {
	for _, param := range ei.csiParam {
		p, err := strconv.Atoi(param)
		if err != nil {
			return errCSIParseError
		}

		switch {
		case p >= 30 && p <= 37:
			ei.curFgColor |= Attribute(p - 30 + 1)
		case p == 39:
			ei.curFgColor |= ColorDefault
		case p >= 40 && p <= 47:
			ei.curBgColor |= Attribute(p - 40 + 1)
		case p == 49:
			ei.curBgColor |= ColorDefault
		case p == 1:
			ei.curFgColor |= AttrBold
		case p == 4:
			ei.curFgColor |= AttrUnderline
		case p == 7:
			ei.curFgColor |= AttrReverse
		case p == 0:
			ei.curFgColor = ColorDefault
			ei.curBgColor = ColorDefault
		}
	}

	return nil
}

// output256 allows you to leverage the 256-colors terminal mode:
//   0x01 - 0x08: the 8 colors as in OutputNormal
//   0x09 - 0x10: Color* | AttrBold
//   0x11 - 0xe8: 216 different colors
//   0xe9 - 0x1ff: 24 different shades of grey
func (ei *escapeInterpreter) output256() error {
	if len(ei.csiParam) < 3 {
		return ei.outputNormal()
	}

	mode, err := strconv.Atoi(ei.csiParam[1])
	if err != nil {
		return errCSIParseError
	}
	if mode != 5 {
		return ei.outputNormal()
	}

	fgbg, err := strconv.Atoi(ei.csiParam[0])
	if err != nil {
		return errCSIParseError
	}
	color, err := strconv.Atoi(ei.csiParam[2])
	if err != nil {
		return errCSIParseError
	}

	switch fgbg {
	case 38:
		ei.curFgColor = Attribute(color + 1)

		for _, param := range ei.csiParam[3:] {
			p, err := strconv.Atoi(param)
			if err != nil {
				return errCSIParseError
			}

			switch {
			case p == 1:
				ei.curFgColor |= AttrBold
			case p == 4:
				ei.curFgColor |= AttrUnderline
			case p == 7:
				ei.curFgColor |= AttrReverse
			}
		}
	case 48:
		ei.curBgColor = Attribute(color + 1)
	default:
		return errCSIParseError
	}

	return nil
}
