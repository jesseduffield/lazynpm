package gui

import "github.com/jesseduffield/lazynpm/pkg/theme"

func (gui *Gui) newStringTask(viewName string, str string) error {
	gui.renderString(viewName, str)
	return nil
}

func (gui *Gui) newMainCommand(cmdStr string, contextKey string) error {
	cmd := gui.OSCommand.ExecutableFromString(cmdStr)

	mainPanelLeft, mainPanelTop, mainPanelRight, mainPanelBottom, err := gui.getMainViewDimensions()
	if err != nil {
		return err
	}

	v, err := gui.g.SetView(contextKey, mainPanelLeft, mainPanelTop, mainPanelRight, mainPanelBottom, 0)
	if err != nil {
		if err.Error() != "unknown view" {
			return err
		}
		v.Wrap = true
		v.FgColor = theme.GocuiDefaultTextColor
		v.Autoscroll = true
	}

	if _, err := gui.g.SetViewOnTop(contextKey); err != nil {
		return err
	}

	gui.State.ContextViews[contextKey] = v

	if err := gui.newPtyTask(contextKey, cmd, cmdStr); err != nil {
		gui.Log.Error(err)
	}
	return nil
}
