package gui

import (
	"fmt"

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
		displayStrings := presentation.GetPackageListDisplayStrings(gui.State.Packages, gui.State.Deps)
		gui.renderDisplayStrings(packagesView, displayStrings)

		displayStrings = presentation.GetDependencyListDisplayStrings(gui.State.Deps)
		gui.renderDisplayStrings(gui.getDepsView(), displayStrings)

		displayStrings = presentation.GetScriptListDisplayStrings(gui.getScripts())
		gui.renderDisplayStrings(gui.getScriptsView(), displayStrings)
		return nil
	})

	gui.refreshStatus()

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

	gui.State.Deps, err = gui.NpmManager.GetDeps(gui.currentPackage())
	if err != nil {
		return err
	}

	gui.refreshSelectedLine(&gui.State.Panels.Packages.SelectedLine, len(gui.State.Packages))
	return nil
}

func (gui *Gui) handleCheckoutPackage(g *gocui.Gui, v *gocui.View) error {
	selectedPkg := gui.getSelectedPackage()

	if selectedPkg == nil {
		return nil
	}

	if err := gui.sendPackageToTop(selectedPkg.Path); err != nil {
		return err
	}

	gui.State.Panels.Packages.SelectedLine = 0
	gui.State.Panels.Deps.SelectedLine = 0
	gui.State.Panels.Scripts.SelectedLine = 0

	return gui.refreshPackages()
}

func (gui *Gui) handleLinkPackage(g *gocui.Gui, v *gocui.View) error {
	// if it's the current package we should globally link it, otherwise we should link it to here
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	currentPkg := gui.currentPackage()
	if currentPkg == nil {
		return nil
	}

	var cmdStr string
	if selectedPkg == currentPkg {
		if selectedPkg.LinkedGlobally {
			cmdStr = "npm unlink"
		} else {
			cmdStr = "npm link"
		}
	} else {
		if gui.linkPathMap()[selectedPkg.Path] {
			cmdStr = fmt.Sprintf("npm unlink --no-save %s", selectedPkg.Config.Name)
		} else {
			cmdStr = fmt.Sprintf("npm link %s", selectedPkg.Config.Name)
		}
	}

	cmd := gui.OSCommand.ExecutableFromString(cmdStr)
	if err := gui.newPtyTask("main", cmd, cmdStr); err != nil {
		gui.Log.Error(err)
	}

	return nil
}

func (gui *Gui) handleInstall() error {
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	currentPkg := gui.currentPackage()
	if currentPkg == nil {
		return nil
	}

	var cmdStr string
	if selectedPkg == currentPkg {
		cmdStr = "npm install"
	} else {
		cmdStr = "npm install --prefix " + selectedPkg.Path
	}

	cmd := gui.OSCommand.ExecutableFromString(cmdStr)
	if err := gui.newPtyTask("main", cmd, cmdStr); err != nil {
		gui.Log.Error(err)
	}
	return nil
}

func (gui *Gui) handleBuild() error {
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	currentPkg := gui.currentPackage()
	if currentPkg == nil {
		return nil
	}

	var cmdStr string
	if selectedPkg == currentPkg {
		cmdStr = "npm run build"
	} else {
		cmdStr = "npm run build --prefix " + selectedPkg.Path
	}

	cmd := gui.OSCommand.ExecutableFromString(cmdStr)
	if err := gui.newPtyTask("main", cmd, cmdStr); err != nil {
		gui.Log.Error(err)
	}
	return nil
}
