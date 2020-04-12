package gui

import (
	"fmt"

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
	return gui.currentPackage().SortedScripts()
}

func (gui *Gui) handleScriptSelect(g *gocui.Gui, v *gocui.View) error {
	script := gui.getSelectedScript()
	if script == nil {
		gui.printToMain(gui.Tr.SLocalize("NoScripts"))
		return nil
	}
	gui.renderString("secondary", presentation.ScriptSummary(script))
	gui.activateContextView(script.ID())
	return nil
}

func (gui *Gui) handleRunScript() error {
	script := gui.getSelectedScript()

	return gui.createPromptPanel(gui.getScriptsView(), "run script", fmt.Sprintf("npm run %s", script.Name), func(input string) error {
		return gui.newMainCommand(input, script.ID())
	})
}

func (gui *Gui) handleRemoveScript() error {
	script := gui.getSelectedScript()

	return gui.createConfirmationPanel(createConfirmationPanelOpts{
		returnToView:       gui.getScriptsView(),
		title:              "Remove script",
		prompt:             fmt.Sprintf("are you sure you want to remove script `%s`?", script.Name),
		returnFocusOnClose: true,
		handleConfirm: func() error {
			return gui.surfaceError(
				gui.NpmManager.RemoveScript(script.Name, gui.currentPackage().ConfigPath()),
			)
		},
	})
}

func (gui *Gui) selectedScriptID() string {
	script := gui.getSelectedScript()
	if script == nil {
		return ""
	}

	return script.ID()
}
