package gui

import (
	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/theme"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

// getFocusLayout returns a manager function for when view gain and lose focus
func (gui *Gui) getFocusLayout() func(g *gocui.Gui) error {
	var previousView *gocui.View
	return func(g *gocui.Gui) error {
		newView := gui.g.CurrentView()
		if err := gui.onFocusChange(); err != nil {
			return err
		}
		// for now we don't consider losing focus to a popup panel as actually losing focus
		if newView != previousView && !gui.isPopupPanel(newView.Name()) {
			if err := gui.onFocusLost(previousView, newView); err != nil {
				return err
			}
			if err := gui.onFocus(newView); err != nil {
				return err
			}
			previousView = newView
		}
		return nil
	}
}

func (gui *Gui) onFocusChange() error {
	currentView := gui.g.CurrentView()
	for _, view := range gui.g.Views() {
		view.Highlight = view == currentView
	}
	return nil
}

func (gui *Gui) onFocusLost(v *gocui.View, newView *gocui.View) error {
	if v == nil {
		return nil
	}
	if v.IsSearching() && newView.Name() != "search" {
		if err := gui.onSearchEscape(); err != nil {
			return err
		}
	}
	switch v.Name() {
	case "main":
		// if we have lost focus to a first-class panel, we need to do some cleanup
		gui.changeMainViewsContext("normal")
	}
	gui.Log.Info(v.Name() + " focus lost")
	return nil
}

func (gui *Gui) onFocus(v *gocui.View) error {
	if v == nil {
		return nil
	}
	gui.Log.Info(v.Name() + " focus gained")
	return nil
}

func (gui *Gui) getViewHeights() map[string]int {
	currView := gui.g.CurrentView()
	currentCyclebleView := gui.State.PreviousView
	if currView != nil {
		viewName := currView.Name()
		usePreviousView := true
		for _, view := range cyclableViews {
			if view == viewName {
				currentCyclebleView = viewName
				usePreviousView = false
				break
			}
		}
		if usePreviousView {
			currentCyclebleView = gui.State.PreviousView
		}
	}

	_, height := gui.g.Size()

	if gui.State.ScreenMode == SCREEN_FULL || gui.State.ScreenMode == SCREEN_HALF {
		vHeights := map[string]int{
			"status":   0,
			"packages": 0,
			"deps":     0,
			"scripts":  0,
			"options":  0,
		}
		vHeights[currentCyclebleView] = height - 1
		return vHeights
	}

	usableSpace := height - 4
	extraSpace := usableSpace - (usableSpace/3)*3

	if height >= 28 {
		return map[string]int{
			"status":   3,
			"packages": (usableSpace / 3) + extraSpace,
			"deps":     usableSpace / 3,
			"scripts":  usableSpace / 3,
			"options":  1,
		}
	}

	defaultHeight := 3
	if height < 21 {
		defaultHeight = 1
	}
	vHeights := map[string]int{
		"status":   defaultHeight,
		"packages": defaultHeight,
		"deps":     defaultHeight,
		"scripts":  defaultHeight,
		"options":  defaultHeight,
	}
	sideViewCount := 3
	vHeights[currentCyclebleView] = height - defaultHeight*sideViewCount - 1

	return vHeights
}

func (gui *Gui) getMainViewDimensions() (int, int, int, int, error) {
	width, height := gui.g.Size()

	sidePanelWidthRatio := gui.Config.GetUserConfig().GetFloat64("gui.sidePanelWidth")

	var leftSideWidth int
	switch gui.State.ScreenMode {
	case SCREEN_NORMAL:
		leftSideWidth = int(float64(width) * sidePanelWidthRatio)
	case SCREEN_HALF:
		leftSideWidth = width/2 - 2
	case SCREEN_FULL:
		currentView := gui.g.CurrentView()
		if currentView != nil && currentView.Name() == "main" {
			leftSideWidth = 0
		} else {
			leftSideWidth = width - 1
		}
	}

	mainPanelLeft := leftSideWidth + 1
	mainPanelRight := width - 1
	mainPanelTop := 6
	secondaryView := gui.getSecondaryView()
	if secondaryView != nil {
		mainPanelTop = len(secondaryView.BufferLines()) + 2
	}
	mainPanelBottom := height - 2

	return mainPanelLeft, mainPanelTop, mainPanelRight, mainPanelBottom, nil
}

// layout is called for every screen re-render e.g. when the screen is resized
func (gui *Gui) layout(g *gocui.Gui) error {
	g.Highlight = true
	width, height := g.Size()

	information := gui.Config.GetVersion()
	if gui.g.Mouse {
		donate := color.New(color.FgMagenta, color.Underline).Sprint(gui.Tr.SLocalize("Donate"))
		information = donate + " " + information
	}

	minimumHeight := 9
	minimumWidth := 10
	if height < minimumHeight || width < minimumWidth {
		v, err := g.SetView("limit", 0, 0, width-1, height-1, 0)
		if err != nil {
			if err.Error() != "unknown view" {
				return err
			}
			v.Title = gui.Tr.SLocalize("NotEnoughSpace")
			v.Wrap = true
			_, _ = g.SetViewOnTop("limit")
		}
		return nil
	}

	vHeights := gui.getViewHeights()

	optionsVersionBoundary := width - max(len(utils.Decolorise(information)), 1)

	appStatus := gui.statusManager.getStatusString()
	appStatusOptionsBoundary := 0
	if appStatus != "" {
		appStatusOptionsBoundary = len(appStatus) + 2
	}

	_, _ = g.SetViewOnBottom("limit")
	_ = g.DeleteView("limit")

	textColor := theme.GocuiDefaultTextColor

	main := "main"
	secondary := "secondary"

	mainPanelLeft, mainPanelTop, mainPanelRight, mainPanelBottom, err := gui.getMainViewDimensions()
	if err != nil {
		return err
	}
	leftSideWidth := mainPanelLeft - 1

	v, err := g.SetView(main, mainPanelLeft, mainPanelTop, mainPanelRight, mainPanelBottom, 0)
	if err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		v.Wrap = true
		v.FgColor = textColor
		v.Autoscroll = true
	}

	for _, view := range gui.State.ContextViews {
		_, _ = g.SetView(view.Name(), mainPanelLeft, mainPanelTop, mainPanelRight, mainPanelBottom, 0)
	}

	hiddenViewOffset := 9999

	secondaryView, err := g.SetView(secondary, mainPanelLeft, 0, width-1, mainPanelTop-1, 0)
	if err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		secondaryView.Wrap = true
		secondaryView.FgColor = gocui.ColorWhite
	}

	if v, err := g.SetView("status", 0, 0, leftSideWidth, vHeights["status"]-1, gocui.BOTTOM|gocui.RIGHT); err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		v.Title = gui.Tr.SLocalize("StatusTitle")
		v.FgColor = textColor
	}

	packagesView, err := g.SetViewBeneath("packages", "status", vHeights["packages"])
	if err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		packagesView.Highlight = true
		packagesView.Title = gui.Tr.SLocalize("PackagesTitle")
		packagesView.ContainsList = true
	}

	depsView, err := g.SetViewBeneath("deps", "packages", vHeights["deps"])
	if err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		depsView.Title = gui.Tr.SLocalize("DepsTitle")
		depsView.FgColor = textColor
		depsView.ContainsList = true
	}

	scriptsView, err := g.SetViewBeneath("scripts", "deps", vHeights["scripts"])
	if err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		scriptsView.Title = gui.Tr.SLocalize("ScriptsTitle")
		scriptsView.FgColor = textColor
		scriptsView.ContainsList = true
	}

	if v, err := g.SetView("options", appStatusOptionsBoundary-1, height-2, optionsVersionBoundary-1, height, 0); err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		v.Frame = false
		v.FgColor = theme.OptionsColor
	}

	searchViewOffset := hiddenViewOffset
	if gui.State.Searching.isSearching {
		searchViewOffset = 0
	}

	// this view takes up one character. Its only purpose is to show the slash when searching
	searchPrefix := "search: "
	if searchPrefixView, err := g.SetView("searchPrefix", appStatusOptionsBoundary-1+searchViewOffset, height-2+searchViewOffset, len(searchPrefix)+searchViewOffset, height+searchViewOffset, 0); err != nil {
		if err.Error() != "unknown view" {
			return err
		}

		searchPrefixView.BgColor = gocui.ColorDefault
		searchPrefixView.FgColor = gocui.ColorGreen
		searchPrefixView.Frame = false
		gui.setViewContent(gui.g, searchPrefixView, searchPrefix)
	}

	if searchView, err := g.SetView("search", appStatusOptionsBoundary-1+searchViewOffset+len(searchPrefix), height-2+searchViewOffset, optionsVersionBoundary+searchViewOffset, height+searchViewOffset, 0); err != nil {
		if err.Error() != "unknown view" {
			return err
		}

		searchView.BgColor = gocui.ColorDefault
		searchView.FgColor = gocui.ColorGreen
		searchView.Frame = false
		searchView.Editable = true
	}

	if appStatusView, err := g.SetView("appStatus", -1, height-2, width, height, 0); err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		appStatusView.BgColor = gocui.ColorDefault
		appStatusView.FgColor = gocui.ColorCyan
		appStatusView.Frame = false
		if _, err := g.SetViewOnBottom("appStatus"); err != nil {
			return err
		}
	}

	informationView, err := g.SetView("information", optionsVersionBoundary-1, height-2, width, height, 0)
	if err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		informationView.BgColor = gocui.ColorDefault
		informationView.FgColor = gocui.ColorGreen
		informationView.Frame = false
		gui.renderString("information", information)

		// doing this here because it'll only happen once
		if err := gui.onInitialViewsCreation(); err != nil {
			return err
		}
	}
	if gui.State.OldInformation != information {
		gui.setViewContent(g, informationView, information)
		gui.State.OldInformation = information
	}

	if gui.g.CurrentView() == nil {
		initialView := gui.getPackagesView()
		if _, err := gui.g.SetCurrentView(initialView.Name()); err != nil {
			return err
		}

		if err := gui.switchFocus(gui.g, nil, initialView); err != nil {
			return err
		}
	}

	type listViewState struct {
		selectedLine int
		lineCount    int
		view         *gocui.View
		context      string
		listView     *listView
	}

	listViewStates := []listViewState{
		{view: packagesView, context: "", selectedLine: gui.State.Panels.Packages.SelectedLine, lineCount: len(gui.State.Packages), listView: gui.packagesListView()},
		{view: depsView, context: "", selectedLine: gui.State.Panels.Deps.SelectedLine, lineCount: len(gui.State.Deps), listView: gui.depsListView()},
		{view: scriptsView, context: "", selectedLine: gui.State.Panels.Scripts.SelectedLine, lineCount: len(gui.getScripts()), listView: gui.scriptsListView()},
	}

	// menu view might not exist so we check to be safe
	if menuView, err := gui.g.View("menu"); err == nil {
		listViewStates = append(listViewStates, listViewState{view: menuView, context: "", selectedLine: gui.State.Panels.Menu.SelectedLine, lineCount: gui.State.MenuItemCount, listView: gui.menuListView()})
	}
	for _, listViewState := range listViewStates {
		// ignore views where the context doesn't match up with the selected line we're trying to focus
		if listViewState.context != "" && (listViewState.view.Context != listViewState.context) {
			continue
		}
		// check if the selected line is now out of view and if so refocus it
		listViewState.view.FocusPoint(0, listViewState.selectedLine)

		// I doubt this is expensive though it's admittedly redundant after the first render
		listViewState.view.SetOnSelectItem(gui.onSelectItemWrapper(listViewState.listView.onSearchSelect))
	}

	mainViewWidth, mainViewHeight := gui.getMainView().Size()
	if mainViewWidth != gui.State.PrevMainWidth || mainViewHeight != gui.State.PrevMainHeight {
		gui.State.PrevMainWidth = mainViewWidth
		gui.State.PrevMainHeight = mainViewHeight
		if err := gui.onResize(); err != nil {
			return err
		}
	}

	// here is a good place log some stuff
	// if you download humanlog and do tail -f development.log | humanlog
	// this will let you see these branches as prettified json
	// gui.Log.Info(utils.AsJson(gui.State.Branches[0:4]))
	return gui.resizeCurrentPopupPanel(g)
}

func (gui *Gui) onInitialViewsCreation() error {
	gui.changeMainViewsContext("normal")

	// here is where you would set initial contexts for views with tabs

	return gui.loadNewRepo()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
