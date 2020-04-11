package gui

import "github.com/jesseduffield/gocui"

type listView struct {
	viewName                string
	context                 string
	getItemsLength          func() int
	getSelectedLineIdxPtr   func() *int
	handleFocus             func(g *gocui.Gui, v *gocui.View) error
	handleItemSelect        func(g *gocui.Gui, v *gocui.View) error
	handleClickSelectedItem func(g *gocui.Gui, v *gocui.View) error
	gui                     *Gui
	rendersToMainView       bool
}

func (lv *listView) handlePrevLine(g *gocui.Gui, v *gocui.View) error {
	return lv.handleLineChange(-1)
}

func (lv *listView) handleNextLine(g *gocui.Gui, v *gocui.View) error {
	return lv.handleLineChange(1)
}

func (lv *listView) handleLineChange(change int) error {
	if !lv.gui.isPopupPanel(lv.viewName) && lv.gui.popupPanelFocused() {
		return nil
	}

	view, err := lv.gui.g.View(lv.viewName)
	if err != nil {
		return err
	}

	lv.gui.changeSelectedLine(lv.getSelectedLineIdxPtr(), lv.getItemsLength(), change)
	view.FocusPoint(0, *lv.getSelectedLineIdxPtr())

	if lv.rendersToMainView {
		if err := lv.gui.resetOrigin(lv.gui.getMainView()); err != nil {
			return err
		}
		if err := lv.gui.resetOrigin(lv.gui.getSecondaryView()); err != nil {
			return err
		}
	}

	return lv.handleItemSelect(lv.gui.g, view)
}

func (lv *listView) handleNextPage(g *gocui.Gui, v *gocui.View) error {
	view, err := lv.gui.g.View(lv.viewName)
	if err != nil {
		return nil
	}
	_, height := view.Size()
	delta := height - 1
	if delta == 0 {
		delta = 1
	}
	return lv.handleLineChange(delta)
}

func (lv *listView) handleGotoTop(g *gocui.Gui, v *gocui.View) error {
	return lv.handleLineChange(-lv.getItemsLength())
}

func (lv *listView) handleGotoBottom(g *gocui.Gui, v *gocui.View) error {
	return lv.handleLineChange(lv.getItemsLength())
}

func (lv *listView) handlePrevPage(g *gocui.Gui, v *gocui.View) error {
	view, err := lv.gui.g.View(lv.viewName)
	if err != nil {
		return nil
	}
	_, height := view.Size()
	delta := height - 1
	if delta == 0 {
		delta = 1
	}
	return lv.handleLineChange(-delta)
}

func (lv *listView) handleClick(g *gocui.Gui, v *gocui.View) error {
	if !lv.gui.isPopupPanel(lv.viewName) && lv.gui.popupPanelFocused() {
		return nil
	}

	selectedLineIdxPtr := lv.getSelectedLineIdxPtr()
	prevSelectedLineIdx := *selectedLineIdxPtr
	newSelectedLineIdx := v.SelectedLineIdx()

	if newSelectedLineIdx > lv.getItemsLength()-1 {
		return lv.handleFocus(lv.gui.g, v)
	}

	*selectedLineIdxPtr = newSelectedLineIdx

	if lv.rendersToMainView {
		if err := lv.gui.resetOrigin(lv.gui.getMainView()); err != nil {
			return err
		}
		if err := lv.gui.resetOrigin(lv.gui.getSecondaryView()); err != nil {
			return err
		}
	}

	prevViewName := lv.gui.currentViewName()
	if prevSelectedLineIdx == newSelectedLineIdx && prevViewName == lv.viewName && lv.handleClickSelectedItem != nil {
		return lv.handleClickSelectedItem(lv.gui.g, v)
	}
	return lv.handleItemSelect(lv.gui.g, v)
}

func (lv *listView) onSearchSelect(selectedLineIdx int) error {
	view, err := lv.gui.g.View(lv.viewName)
	if err != nil {
		return nil
	}

	*lv.getSelectedLineIdxPtr() = selectedLineIdx
	return lv.handleItemSelect(lv.gui.g, view)
}

func (gui *Gui) menuListView() *listView {
	return &listView{
		viewName:              "menu",
		getItemsLength:        func() int { return gui.getMenuView().LinesHeight() },
		getSelectedLineIdxPtr: func() *int { return &gui.State.Panels.Menu.SelectedLine },
		handleFocus:           gui.handleMenuSelect,
		handleItemSelect:      gui.handleMenuSelect,
		// need to add a layer of indirection here because the callback changes during runtime
		handleClickSelectedItem: gui.wrappedHandler(func() error { return gui.State.Panels.Menu.OnPress(gui.g, nil) }),
		gui:                     gui,
		rendersToMainView:       false,
	}
}

func (gui *Gui) packagesListView() *listView {
	return &listView{
		viewName:              "packages",
		getItemsLength:        func() int { return len(gui.State.Packages) },
		getSelectedLineIdxPtr: func() *int { return &gui.State.Panels.Packages.SelectedLine },
		handleFocus:           gui.handlePackageSelect,
		handleItemSelect:      gui.handlePackageSelect,
		gui:                   gui,
		rendersToMainView:     true,
	}
}

func (gui *Gui) depsListView() *listView {
	return &listView{
		viewName: "deps",
		// TODO: handle more dep types
		getItemsLength:        func() int { return len(gui.currentPackage().SortedDependencies()) },
		getSelectedLineIdxPtr: func() *int { return &gui.State.Panels.Deps.SelectedLine },
		handleFocus:           gui.handleDepSelect,
		handleItemSelect:      gui.handleDepSelect,
		gui:                   gui,
		rendersToMainView:     true,
	}
}

func (gui *Gui) scriptsListView() *listView {
	return &listView{
		viewName:              "scripts",
		getItemsLength:        func() int { return len(gui.getScripts()) },
		getSelectedLineIdxPtr: func() *int { return &gui.State.Panels.Scripts.SelectedLine },
		handleFocus:           gui.handleScriptSelect,
		handleItemSelect:      gui.handleScriptSelect,
		gui:                   gui,
		rendersToMainView:     true,
	}
}

func (gui *Gui) getListViews() []*listView {
	return []*listView{
		gui.menuListView(),
		gui.packagesListView(),
		gui.depsListView(),
		gui.scriptsListView(),
	}
}
