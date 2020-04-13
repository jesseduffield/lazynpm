package gui

import (
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
