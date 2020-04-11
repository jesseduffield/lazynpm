package gui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/gui/presentation"
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

func (gui *Gui) getScripts() []*commands.Script {
	currentPackage := gui.currentPackage()
	if currentPackage == nil {
		return nil
	}

	return currentPackage.SortedScripts()
}

func (gui *Gui) handleScriptSelect(g *gocui.Gui, v *gocui.View) error {
	script := gui.getSelectedScript()
	if script == nil {
		gui.getMainView().Title = ""
		return gui.newStringTask("main", gui.Tr.SLocalize("NoScripts"))
	}
	gui.renderString("secondary", presentation.ScriptSummary(script))
	return nil
}

func (gui *Gui) handleRunScript() error {
	script := gui.getSelectedScript()

	return gui.createPromptPanel(gui.getScriptsView(), "run script", fmt.Sprintf("npm run %s ", script.Name), func(g *gocui.Gui, v *gocui.View) error {
		cmdStr := strings.TrimSpace(v.Buffer())
		cmd := gui.OSCommand.ExecutableFromString(cmdStr)
		if err := gui.newPtyTask("main", cmd, cmdStr); err != nil {
			gui.Log.Error(err)
		}
		return nil
	})
}
