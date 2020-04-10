package gui

import (
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
	pkg := gui.getSelectedPackage()
	if pkg == nil {
		gui.getMainView().Title = ""
		return gui.newStringTask("main", gui.Tr.SLocalize("NoChangedPackages"))
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

		displayStrings = presentation.GetDependencyListDisplayStrings(gui.currentPackage().SortedDeps())
		gui.renderDisplayStrings(gui.getDepsView(), displayStrings)

		displayStrings = presentation.GetScriptListDisplayStrings(gui.currentPackage().SortedScripts())
		gui.renderDisplayStrings(gui.getScriptsView(), displayStrings)
		return nil
	})

	return nil
}

func (gui *Gui) currentPackage() *commands.Package {
	if len(gui.State.Packages) == 0 {
		panic("need at least one package")
	}
	return gui.State.Packages[0]
}

// specific functions

func (gui *Gui) refreshStatePackages() error {
	// get files to stage
	var err error
	gui.State.Packages, err = gui.NpmManager.GetPackages(gui.Config.GetAppState().RecentPackages)
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
