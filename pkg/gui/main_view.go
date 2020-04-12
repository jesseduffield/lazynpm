package gui

func (gui *Gui) handleEscapeMain() error {
	viewName := gui.State.CurrentSideView
	view, err := gui.g.View(viewName)
	if err != nil {
		return nil
	}
	return gui.switchFocus(gui.g, nil, view)
}
