package gui

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/theme"
)

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

		bindings := []*Binding{
			{
				ViewName:    contextKey,
				Key:         gocui.MouseWheelDown,
				Modifier:    gocui.ModNone,
				Handler:     gui.scrollDownMain,
				Description: gui.Tr.SLocalize("ScrollDown"),
				Alternative: "fn+up",
			},
			{
				ViewName:    contextKey,
				Key:         gocui.MouseWheelUp,
				Modifier:    gocui.ModNone,
				Handler:     gui.scrollUpMain,
				Description: gui.Tr.SLocalize("ScrollUp"),
				Alternative: "fn+down",
			},
			{
				ViewName: contextKey,
				Key:      gocui.MouseLeft,
				Modifier: gocui.ModNone,
				Handler:  gui.handleMouseDownMain,
			},
			{
				ViewName: contextKey,
				Key:      gui.getKey("universal.return"),
				Modifier: gocui.ModNone,
				Handler:  gui.wrappedHandler(gui.handleEscapeMain),
			},
		}

		for _, binding := range bindings {
			if err := gui.g.SetKeybinding(binding.ViewName, nil, binding.Key, binding.Modifier, binding.Handler); err != nil {
				return err
			}
		}
	}

	if _, err := gui.g.SetViewOnTop(contextKey); err != nil {
		return err
	}

	gui.State.CommandMap[contextKey] = &commands.CommandView{
		View: v,
		Cmd:  cmd,
	}

	if err := gui.newPtyTask(contextKey, cmd, cmdStr); err != nil {
		gui.Log.Error(err)
	}

	// we need to refresh packages to show that a command is now in flight
	return gui.refreshPackages()
}
