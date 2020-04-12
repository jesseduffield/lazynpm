package gui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/utils"
	"github.com/spkg/bom"
)

var cyclableViews = []string{"status", "packages", "deps", "scripts"}

func intArrToMap(arr []int) map[int]bool {
	output := map[int]bool{}
	for _, el := range arr {
		output[el] = true
	}
	return output
}

const (
	ASYNC = iota
	SYNC
	BLOCK_UI
)

type refreshOptions struct {
	mode int
}

func (gui *Gui) refreshSidePanels(options refreshOptions) error {
	// refresh status, refresh dependencies, refresh
	return nil
}

func (gui *Gui) nextView(g *gocui.Gui, v *gocui.View) error {
	var focusedViewName string
	if v == nil || v.Name() == cyclableViews[len(cyclableViews)-1] {
		focusedViewName = cyclableViews[0]
	} else {
		// if we're in the commitFiles view we'll act like we're in the commits view
		viewName := v.Name()
		for i := range cyclableViews {
			if viewName == cyclableViews[i] {
				focusedViewName = cyclableViews[i+1]
				break
			}
			if i == len(cyclableViews)-1 {
				return nil
			}
		}
	}
	focusedView, err := g.View(focusedViewName)
	if err != nil {
		panic(err)
	}
	if err := gui.resetOrigin(gui.getMainView()); err != nil {
		return err
	}
	return gui.switchFocus(g, v, focusedView)
}

func (gui *Gui) previousView(g *gocui.Gui, v *gocui.View) error {
	var focusedViewName string
	if v == nil || v.Name() == cyclableViews[0] {
		focusedViewName = cyclableViews[len(cyclableViews)-1]
	} else {
		viewName := v.Name()
		for i := range cyclableViews {
			if viewName == cyclableViews[i] {
				focusedViewName = cyclableViews[i-1] // TODO: make this work properly
				break
			}
			if i == len(cyclableViews)-1 {
				return nil
			}
		}
	}
	focusedView, err := g.View(focusedViewName)
	if err != nil {
		panic(err)
	}
	if err := gui.resetOrigin(gui.getMainView()); err != nil {
		return err
	}
	return gui.switchFocus(g, v, focusedView)
}

func (gui *Gui) newLineFocused(g *gocui.Gui, v *gocui.View) error {
	switch v.Name() {
	case "menu":
		return gui.handleMenuSelect(g, v)
	case "status":
		return gui.handleStatusSelect(g, v)
	case "packages":
		return gui.handlePackageSelect(g, v)
	case "deps":
		return gui.handleDepSelect(g, v)
	case "scripts":
		return gui.handleScriptSelect(g, v)
	case "main":
		v.Highlight = false
		return nil
	case "search", "confirmation":
		return nil
	default:
		panic(gui.Tr.SLocalize("NoViewMachingNewLineFocusedSwitchStatement"))
	}
}

func (gui *Gui) returnFocus(g *gocui.Gui, v *gocui.View) error {
	previousView, err := g.View(gui.State.PreviousView)
	if err != nil {
		// always fall back to files view if there's no 'previous' view stored
		previousView, err = g.View("files")
		if err != nil {
			gui.Log.Error(err)
		}
	}
	return gui.switchFocus(g, v, previousView)
}

func (gui *Gui) goToSideView(sideViewName string) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		view, err := g.View(sideViewName)
		if err != nil {
			gui.Log.Error(err)
			return nil
		}
		err = gui.closePopupPanels()
		if err != nil {
			gui.Log.Error(err)
			return nil
		}
		return gui.switchFocus(g, nil, view)
	}
}

func (gui *Gui) closePopupPanels() error {
	gui.onNewPopupPanel()
	err := gui.closeConfirmationPrompt(gui.g, true)
	if err != nil {
		gui.Log.Error(err)
		return err
	}
	return nil
}

// pass in oldView = nil if you don't want to be able to return to your old view
// TODO: move some of this logic into our onFocusLost and onFocus hooks
func (gui *Gui) switchFocus(g *gocui.Gui, oldView, newView *gocui.View) error {
	// we assume we'll never want to return focus to a popup panel i.e.
	// we should never stack popup panels
	if oldView != nil && !gui.isPopupPanel(oldView.Name()) {
		gui.State.PreviousView = oldView.Name()
	}

	if utils.IncludesString(cyclableViews, newView.Name()) {
		gui.State.CurrentSideView = newView.Name()
	}

	if _, err := g.SetCurrentView(newView.Name()); err != nil {
		return err
	}
	if _, err := g.SetViewOnTop(newView.Name()); err != nil {
		return err
	}

	g.Cursor = newView.Editable

	if err := gui.renderPanelOptions(); err != nil {
		return err
	}

	return gui.newLineFocused(g, newView)
}

func (gui *Gui) resetOrigin(v *gocui.View) error {
	_ = v.SetCursor(0, 0)
	return v.SetOrigin(0, 0)
}

func (gui *Gui) cleanString(s string) string {
	output := string(bom.Clean([]byte(s)))
	return utils.NormalizeLinefeeds(output)
}

func (gui *Gui) setViewContent(g *gocui.Gui, v *gocui.View, s string) {
	v.Clear()
	fmt.Fprint(v, gui.cleanString(s))
}

// renderString resets the origin of a view and sets its content
func (gui *Gui) renderString(viewName, s string) {
	gui.g.Update(func(*gocui.Gui) error {
		v, err := gui.g.View(viewName)
		if err != nil {
			return nil // return gracefully if view has been deleted
		}
		if err := v.SetOrigin(0, 0); err != nil {
			return err
		}
		if err := v.SetCursor(0, 0); err != nil {
			return err
		}
		gui.setViewContent(gui.g, v, s)
		return nil
	})
}

func (gui *Gui) optionsMapToString(optionsMap map[string]string) string {
	optionsArray := make([]string, 0)
	for key, description := range optionsMap {
		optionsArray = append(optionsArray, key+": "+description)
	}
	sort.Strings(optionsArray)
	return strings.Join(optionsArray, ", ")
}

func (gui *Gui) renderOptionsMap(optionsMap map[string]string) error {
	gui.renderString("options", gui.optionsMapToString(optionsMap))
	return nil
}

func (gui *Gui) getPackagesView() *gocui.View {
	v, _ := gui.g.View("packages")
	return v
}

func (gui *Gui) getDepsView() *gocui.View {
	v, _ := gui.g.View("deps")
	return v
}

func (gui *Gui) getScriptsView() *gocui.View {
	v, _ := gui.g.View("scripts")
	return v
}

func (gui *Gui) getMainView() *gocui.View {
	v, _ := gui.g.View("main")
	return v
}

func (gui *Gui) getSecondaryView() *gocui.View {
	v, _ := gui.g.View("secondary")
	return v
}

func (gui *Gui) getMenuView() *gocui.View {
	v, _ := gui.g.View("menu")
	return v
}

func (gui *Gui) getSearchView() *gocui.View {
	v, _ := gui.g.View("search")
	return v
}

func (gui *Gui) getStatusView() *gocui.View {
	v, _ := gui.g.View("status")
	return v
}

func (gui *Gui) trimmedContent(v *gocui.View) string {
	return strings.TrimSpace(v.Buffer())
}

func (gui *Gui) currentViewName() string {
	currentView := gui.g.CurrentView()
	return currentView.Name()
}

func (gui *Gui) resizeCurrentPopupPanel(g *gocui.Gui) error {
	v := g.CurrentView()
	if gui.isPopupPanel(v.Name()) {
		return gui.resizePopupPanel(g, v)
	}
	return nil
}

func (gui *Gui) resizePopupPanel(g *gocui.Gui, v *gocui.View) error {
	// If the confirmation panel is already displayed, just resize the width,
	// otherwise continue
	content := v.Buffer()
	x0, y0, x1, y1 := gui.getConfirmationPanelDimensions(g, v.Wrap, content)
	vx0, vy0, vx1, vy1 := v.Dimensions()
	if vx0 == x0 && vy0 == y0 && vx1 == x1 && vy1 == y1 {
		return nil
	}
	gui.Log.Info(gui.Tr.SLocalize("resizingPopupPanel"))
	_, err := g.SetView(v.Name(), x0, y0, x1, y1, 0)
	return err
}

func (gui *Gui) changeSelectedLine(line *int, total int, change int) {
	// TODO: find out why we're doing this
	if *line == -1 {
		return
	}
	if *line+change < 0 {
		*line = 0
	} else if *line+change >= total {
		*line = total - 1
	} else {
		*line += change
	}
}

func (gui *Gui) refreshSelectedLine(line *int, total int) {
	if *line == -1 && total > 0 {
		*line = 0
	} else if total-1 < *line {
		*line = total - 1
	}
}

func (gui *Gui) renderDisplayStrings(v *gocui.View, displayStrings [][]string) {
	gui.g.Update(func(g *gocui.Gui) error {
		list := utils.RenderDisplayStrings(displayStrings)
		v.Clear()
		fmt.Fprint(v, list)
		return nil
	})
}

func (gui *Gui) renderPanelOptions() error {
	currentView := gui.g.CurrentView()
	switch currentView.Name() {
	case "menu":
		return gui.renderMenuOptions()
	}
	return gui.renderGlobalOptions()
}

func (gui *Gui) renderGlobalOptions() error {
	return gui.renderOptionsMap(map[string]string{
		fmt.Sprintf("%s/%s", gui.getKeyDisplay("universal.scrollUpMain"), gui.getKeyDisplay("universal.scrollDownMain")):                                                                                 gui.Tr.SLocalize("scroll"),
		fmt.Sprintf("%s %s %s %s", gui.getKeyDisplay("universal.prevBlock"), gui.getKeyDisplay("universal.nextBlock"), gui.getKeyDisplay("universal.prevItem"), gui.getKeyDisplay("universal.nextItem")): gui.Tr.SLocalize("navigate"),
		fmt.Sprintf("%s/%s", gui.getKeyDisplay("universal.return"), gui.getKeyDisplay("universal.quit")):                                                                                                 gui.Tr.SLocalize("close"),
		gui.getKeyDisplay("universal.optionMenu"): gui.Tr.SLocalize("menu"),
		"1-4": gui.Tr.SLocalize("jump"),
	})
}

func (gui *Gui) isPopupPanel(viewName string) bool {
	return viewName == "commitMessage" || viewName == "credentials" || viewName == "confirmation" || viewName == "menu"
}

func (gui *Gui) popupPanelFocused() bool {
	return gui.isPopupPanel(gui.currentViewName())
}

func (gui *Gui) handleClick(v *gocui.View, itemCount int, selectedLine *int, handleSelect func(*gocui.Gui, *gocui.View) error) error {
	if gui.popupPanelFocused() && v != nil && !gui.isPopupPanel(v.Name()) {
		return nil
	}

	if _, err := gui.g.SetCurrentView(v.Name()); err != nil {
		return err
	}

	newSelectedLine := v.SelectedLineIdx()

	if newSelectedLine < 0 {
		newSelectedLine = 0
	}

	if newSelectedLine > itemCount-1 {
		newSelectedLine = itemCount - 1
	}

	*selectedLine = newSelectedLine

	return handleSelect(gui.g, v)
}

// often gocui wants functions in the form `func(g *gocui.Gui, v *gocui.View) error`
// but sometimes we just have a function that returns an error, so this is a
// convenience wrapper to give gocui what it wants.
func (gui *Gui) wrappedHandler(f func() error) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		return f()
	}
}
