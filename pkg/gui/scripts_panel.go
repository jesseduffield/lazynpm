package gui

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
)

// list panel functions

func (gui *Gui) getSelectedScript() *commands.Script {
	currentPackage := gui.currentPackage()

	scripts := currentPackage.SortedScripts()
	if len(scripts) == 0 {
		return nil
	}
	return scripts[gui.State.Panels.Scripts.SelectedLine]
}

func (gui *Gui) handleScriptSelect(g *gocui.Gui, v *gocui.View) error {
	dep := gui.getSelectedScript()
	if dep == nil {
		gui.getMainView().Title = ""
		return gui.newStringTask("main", gui.Tr.SLocalize("NoScripts"))
	}
	return nil
}
