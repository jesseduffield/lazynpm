// +build !windows

package gui

import (
	"fmt"
	"io"

	"github.com/creack/pty"
	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/theme"
	"github.com/jesseduffield/lazynpm/pkg/utils"
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

		selectedLine := 0

		lv := &listView{
			viewName:              contextKey,
			getItemsLength:        func() int { return len(v.BufferLines()) },
			getSelectedLineIdxPtr: func() *int { return &selectedLine },
			handleItemSelect: gui.wrappedHandler(func() error {
				if selectedLine >= len(v.BufferLines())-1 {
					v.Autoscroll = true
				}
				return nil
			}),
			gui:               gui,
			rendersToMainView: false,
		}

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
			{
				ViewName: contextKey,
				Key:      gui.getKey("universal.return"),
				Handler:  gui.handleSearchEscape,
			},
			{
				ViewName:    contextKey,
				Key:         gui.getKey("universal.startSearch"),
				Handler:     gui.handleOpenSearch,
				Description: gui.Tr.SLocalize("startSearch"),
			},
			{ViewName: contextKey, Key: gui.getKey("universal.prevItem-alt"), Handler: gui.unsetAutoScrollWrapper(gui.scrollUpMain)},
			{ViewName: contextKey, Key: gui.getKey("universal.prevItem"), Handler: gui.unsetAutoScrollWrapper(gui.scrollUpMain)},
			{ViewName: contextKey, Key: gui.getKey("universal.nextItem-alt"), Handler: gui.scrollDownMain},
			{ViewName: contextKey, Key: gui.getKey("universal.nextItem"), Handler: gui.scrollDownMain},
			{ViewName: contextKey, Key: gui.getKey("universal.prevPage"), Handler: gui.unsetAutoScrollWrapper(lv.handlePrevPage), Description: gui.Tr.SLocalize("prevPage")},
			{ViewName: contextKey, Key: gui.getKey("universal.nextPage"), Handler: lv.handleNextPage, Description: gui.Tr.SLocalize("nextPage")},
			{ViewName: contextKey, Key: gui.getKey("universal.gotoTop"), Handler: gui.unsetAutoScrollWrapper(lv.handleGotoTop), Description: gui.Tr.SLocalize("gotoTop")},
			{
				ViewName:    contextKey,
				Key:         gui.getKey("universal.gotoBottom"),
				Handler:     lv.handleGotoBottom,
				Description: gui.Tr.SLocalize("gotoBottom"),
			},
		}

		for _, binding := range bindings {
			if err := gui.g.SetKeybinding(binding.ViewName, nil, binding.Key, binding.Modifier, binding.Handler); err != nil {
				return err
			}
		}
	}

	v.SetOnSelectItem(gui.onSelectItemWrapper(func(selectedLineIdx int) error {
		v.FocusPoint(0, selectedLineIdx)
		return nil
	}))

	if _, err := gui.g.SetViewOnTop(contextKey); err != nil {
		return err
	}

	commandView := &commands.CommandView{
		View: v,
		Cmd:  cmd,
	}

	gui.State.CommandViewMap[contextKey] = commandView

	if err := gui.newPtyTask(contextKey, commandView, cmdStr); err != nil {
		gui.Log.Error(err)
	}

	// we need to refresh packages to show that a command is now in flight
	return gui.refreshPackages()
}

func (gui *Gui) unsetAutoScrollWrapper(f func(g *gocui.Gui, v *gocui.View) error) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Autoscroll = false
		return f(g, v)
	}
}

func (gui *Gui) onResize() error {
	if gui.State.Ptmx == nil {
		return nil
	}
	mainView := gui.getMainView()
	width, height := mainView.Size()

	if err := pty.Setsize(gui.State.Ptmx, &pty.Winsize{Cols: uint16(width), Rows: uint16(height)}); err != nil {
		return err
	}

	// TODO: handle resizing properly

	return nil
}

// Some commands need to output for a terminal to active certain behaviour.
// For example,  git won't invoke the GIT_PAGER env var unless it thinks it's
// talking to a terminal. We typically write cmd outputs straight to a view,
// which is just an io.Reader. the pty package lets us wrap a command in a
// pseudo-terminal meaning we'll get the behaviour we want from the underlying
// command.
func (gui *Gui) newPtyTask(viewName string, commandView *commands.CommandView, cmdStr string) error {
	go func() {
		view, err := gui.g.View(viewName)
		if err != nil {
			return // swallowing for now
		}

		view.Clear()

		ptmx, err := pty.Start(commandView.Cmd)
		if err != nil {
			// swallowing for now (actually continue to swallow this)
			return
		}

		// autoscroll might have been turned off if the user scrolled midway through the last command
		view.Autoscroll = true
		view.StdinWriter = ptmx
		view.Pty = true

		gui.State.Ptmx = ptmx

		if err := gui.onResize(); err != nil {
			// swallowing for now
			return
		}

		fmt.Fprint(view, utils.ColoredString(fmt.Sprintf("+ %s\n\n", cmdStr), color.FgYellow))

		_, _ = io.Copy(view, ptmx)

		ptmx.Close()
		gui.State.Ptmx = nil
		view.Pty = false
		view.StdinWriter = nil
		_ = commandView.Cmd.Wait()
		_ = gui.refreshPackages()

		if commandView.Cancelled {
			fmt.Fprint(view, utils.ColoredString("\n\ncommand cancelled", color.FgRed))
		} else if commandView.Cmd.ProcessState.Success() {
			fmt.Fprint(view, utils.ColoredString("\n\ncommand completed successfully", color.FgGreen))
		} else {
			fmt.Fprint(view, utils.ColoredString("\n\ncommand failed", color.FgRed))
		}

	}()
	return nil
}
