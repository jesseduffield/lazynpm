package gui

import (

	// "io"
	// "io/ioutil"

	// "strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/gui/presentation"
)

// list panel functions

func (gui *Gui) getSelectedPackage() *commands.Package {
	if len(gui.State.Packages) == 0 {
		return nil
	}
	return gui.State.Packages[gui.State.Panels.Packages.SelectedLine]
}

func (gui *Gui) handlePackageSelect(g *gocui.Gui, v *gocui.View) error {
	return nil
}

func (gui *Gui) selectPackage() error {
	gui.getPackagesView().FocusPoint(0, gui.State.Panels.Packages.SelectedLine)

	pkg := gui.getSelectedPackage()
	if pkg == nil {
		gui.getMainView().Title = ""
		return gui.newStringTask("main", gui.Tr.SLocalize("NoChangedPackages"))
	}

	if err := gui.resetOrigin(gui.getMainView()); err != nil {
		return err
	}
	if err := gui.resetOrigin(gui.getSecondaryView()); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) refreshPackages() error {
	packagesView := gui.getPackagesView()
	if packagesView == nil {
		// if the filesView hasn't been instantiated yet we just return
		return nil
	}
	if err := gui.refreshStatePackages(); err != nil {
		return err
	}

	gui.g.Update(func(g *gocui.Gui) error {
		displayStrings := presentation.GetPackageListDisplayStrings(gui.State.Packages)
		gui.renderDisplayStrings(packagesView, displayStrings)
		return nil
	})

	return nil
}

// specific functions

func (gui *Gui) refreshStatePackages() error {
	// get files to stage
	var err error
	gui.State.Packages, err = gui.NpmManager.GetPackages()
	if err != nil {
		return err
	}

	gui.refreshSelectedLine(&gui.State.Panels.Packages.SelectedLine, len(gui.State.Packages))
	return nil
}

func (gui *Gui) onPackagesPanelSearchSelect(selectedLine int) error {
	gui.State.Panels.Packages.SelectedLine = selectedLine
	return gui.handlePackageSelect(gui.g, gui.getPackagesView())
}
