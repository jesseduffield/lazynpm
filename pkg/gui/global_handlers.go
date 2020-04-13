package gui

import (
	"os/exec"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

func (gui *Gui) nextScreenMode(g *gocui.Gui, v *gocui.View) error {
	gui.State.ScreenMode = utils.NextIntInCycle([]int{SCREEN_NORMAL, SCREEN_HALF, SCREEN_FULL}, gui.State.ScreenMode)

	return nil
}

func (gui *Gui) prevScreenMode(g *gocui.Gui, v *gocui.View) error {
	gui.State.ScreenMode = utils.PrevIntInCycle([]int{SCREEN_NORMAL, SCREEN_HALF, SCREEN_FULL}, gui.State.ScreenMode)

	return nil
}

func (gui *Gui) scrollUpView(viewName string) error {
	view, _ := gui.g.View(viewName)
	view.Autoscroll = false
	ox, oy := view.Origin()
	scrollHeight := gui.Config.GetUserConfig().GetInt("gui.scrollHeight")
	newOy := oy - scrollHeight
	if newOy <= 0 {
		newOy = 0
	}
	return view.SetOrigin(ox, newOy)
}

func (gui *Gui) scrollDownView(viewName string) error {
	view, _ := gui.g.View(viewName)
	view.Autoscroll = false
	ox, oy := view.Origin()
	_, sy := view.Size()
	y := oy + sy
	scrollHeight := gui.Config.GetUserConfig().GetInt("gui.scrollHeight")
	if y < view.LinesHeight()-1 {
		if err := view.SetOrigin(ox, oy+scrollHeight); err != nil {
			return err
		}
	} else {
		view.Autoscroll = true
	}

	return nil
}

func (gui *Gui) currentContextViewID() string {
	switch gui.State.CurrentSideView {
	case "packages":
		return gui.selectedPackageID()
	case "deps":
		return gui.selectedDepID()
	case "scripts":
		return gui.selectedScriptID()
	case "tarballs":
		return gui.selectedTarballID()
	}
	return ""
}

func (gui *Gui) scrollUpMain(g *gocui.Gui, v *gocui.View) error {
	id := gui.currentContextViewID()
	if id == "" {
		return nil
	}
	return gui.scrollUpView(id)
}

func (gui *Gui) scrollDownMain(g *gocui.Gui, v *gocui.View) error {
	id := gui.currentContextViewID()
	if id == "" {
		return nil
	}
	return gui.scrollDownView(id)
}

func (gui *Gui) scrollUpConfirmationPanel(g *gocui.Gui, v *gocui.View) error {
	if v.Editable {
		return nil
	}
	return gui.scrollUpView("confirmation")
}

func (gui *Gui) scrollDownConfirmationPanel(g *gocui.Gui, v *gocui.View) error {
	if v.Editable {
		return nil
	}
	return gui.scrollDownView("confirmation")
}

func (gui *Gui) handleRefresh(g *gocui.Gui, v *gocui.View) error {
	return gui.refreshPackages()
}

func (gui *Gui) handleMouseDownMain(g *gocui.Gui, v *gocui.View) error {
	if gui.popupPanelFocused() {
		return nil
	}

	id := gui.currentContextViewID()
	view, err := gui.g.View(id)
	if err != nil {
		return nil
	}

	return gui.switchFocus(gui.g.CurrentView(), view)
}

func (gui *Gui) handleInfoClick(g *gocui.Gui, v *gocui.View) error {
	if !gui.g.Mouse {
		return nil
	}

	cx, _ := v.Cursor()

	if cx <= len(gui.Tr.SLocalize("Donate")) {
		return gui.OSCommand.OpenLink("https://github.com/sponsors/jesseduffield")
	}
	return nil
}

func (gui *Gui) editFile(filename string) error {
	_, err := gui.runSyncOrAsyncCommand(gui.OSCommand.EditFile(filename))
	return err
}

func (gui *Gui) openFile(filename string) error {
	if err := gui.OSCommand.OpenFile(filename); err != nil {
		return gui.surfaceError(err)
	}
	return nil
}

// runSyncOrAsyncCommand takes the output of a command that may have returned
// either no error, an error, or a subprocess to execute, and if a subprocess
// needs to be set on the gui object, it does so, and then returns the error
// the bool returned tells us whether the calling code should continue
func (gui *Gui) runSyncOrAsyncCommand(sub *exec.Cmd, err error) (bool, error) {
	if err != nil {
		if err != gui.Errors.ErrSubProcess {
			return false, gui.surfaceError(err)
		}
	}
	if sub != nil {
		gui.SubProcess = sub
		return false, gui.Errors.ErrSubProcess
	}
	return true, nil
}

func (gui *Gui) handleKillCommand() error {
	contextViewId := gui.currentContextViewID()
	if contextViewId == "" {
		return nil
	}

	commandView := gui.State.CommandViewMap[contextViewId]
	if commandView == nil {
		return nil
	}

	commandView.Cancelled = true

	return gui.surfaceError(commands.Kill(commandView.Cmd))
}

func (gui *Gui) finalStep(err error) error {
	if err != nil {
		return gui.createErrorPanel(err.Error())
	}

	return gui.surfaceError(gui.refreshPackages())
}

func (gui *Gui) enterMainView() error {
	id := gui.currentContextViewID()
	view, err := gui.g.View(id)
	if err != nil {
		return nil
	}
	return gui.switchFocus(gui.g.CurrentView(), view)
}
