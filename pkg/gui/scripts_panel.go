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
		gui.getMainView().Title = ""
		return gui.newStringTask("main", gui.Tr.SLocalize("NoScripts"))
	}
	gui.renderString("secondary", presentation.ScriptSummary(script))
	return nil
}

func (gui *Gui) handleRunScript() error {
	script := gui.getSelectedScript()

	return gui.createPromptPanel(gui.getScriptsView(), "run script", fmt.Sprintf("npm run %s ", script.Name), func(input string) error {
		return gui.newMainCommand(input, gui.scriptContextKey(script))
	})
}

func (gui *Gui) scriptContextKey(script *commands.Script) string {
	return fmt.Sprintf("package:%s|script:%s", gui.currentPackage().Path, script.Name)
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
