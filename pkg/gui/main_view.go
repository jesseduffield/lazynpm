package gui

func (gui *Gui) handleEscapeMain() error {
	return gui.returnFocus(gui.g, nil)
}
