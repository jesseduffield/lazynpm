package gui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/gui/presentation"
)

// list panel functions

func (gui *Gui) getSelectedTarball() *commands.Tarball {
	tarballs := gui.State.Tarballs
	if len(tarballs) == 0 {
		return nil
	}
	return tarballs[gui.State.Panels.Tarballs.SelectedLine]
}

func (gui *Gui) handleTarballSelect(g *gocui.Gui, v *gocui.View) error {
	if !gui.showTarballsView() {
		// we hide the tarball view when there are no tarballs
		if err := gui.switchFocus(nil, gui.getScriptsView()); err != nil {
			return err
		}
	}

	tarball := gui.getSelectedTarball()
	if tarball == nil {
		return nil
	}
	gui.renderString("secondary", presentation.TarballSummary(tarball))
	gui.activateContextView(tarball.ID())
	return nil
}

func (gui *Gui) selectedTarballID() string {
	tarball := gui.getSelectedTarball()
	if tarball == nil {
		return ""
	}

	return tarball.ID()
}

func (gui *Gui) wrappedTarballHandler(f func(*commands.Tarball) error) func(*gocui.Gui, *gocui.View) error {
	return gui.wrappedHandler(func() error {
		tarball := gui.getSelectedTarball()
		if tarball == nil {
			return nil
		}

		return gui.finalStep(f(tarball))
	})
}

func (gui *Gui) handleDeleteTarball(tarball *commands.Tarball) error {
	return gui.createConfirmationPanel(createConfirmationPanelOpts{
		returnToView:       gui.getTarballsView(),
		returnFocusOnClose: true,
		title:              "Remove tarball",
		prompt:             fmt.Sprintf("are you sure you want to delete `%s`?", tarball.Name),
		handleConfirm: func() error {
			return gui.finalStep(gui.OSCommand.Remove(tarball.Path))
		},
	})
}

func (gui *Gui) handleInstallTarball(tarball *commands.Tarball) error {
	cmdStr := fmt.Sprintf("npm install %s", tarball.Name)
	return gui.newMainCommand(cmdStr, tarball.ID(), newMainCommandOptions{})
}

func (gui *Gui) handlePublishTarball(tarball *commands.Tarball) error {
	// saying scoped: true because that forces us to specify whether we want to publish
	// as public or restricted. Can't know whether it's a scoped tarball just from
	// the name because the @ is missing
	return gui.handlePublish(tarball.Name, true, tarball.ID())
}

func (gui *Gui) showTarballsView() bool {
	return len(gui.State.Tarballs) > 0
}
