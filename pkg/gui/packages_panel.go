package gui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/go-errors/errors"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/gui/presentation"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

// list panel functions

func (gui *Gui) getSelectedPackage() *commands.Package {
	if len(gui.State.Packages) == 0 {
		return nil
	}
	return gui.State.Packages[gui.State.Panels.Packages.SelectedLine]
}

func (gui *Gui) activateContextView(viewName string) {
	if gui.State.ContextViews[viewName] == nil {
		viewName = "main"
		gui.getMainView().Clear()
	}
	_, _ = gui.g.SetViewOnTop(viewName)
}

func (gui *Gui) printToMain(str string) {
	gui.renderString("main", str)
	_, _ = gui.g.SetViewOnTop("main")
}

func (gui *Gui) handlePackageSelect(g *gocui.Gui, v *gocui.View) error {
	pkg := gui.getSelectedPackage()
	if pkg == nil {
		return nil
	}
	summary := presentation.PackageSummary(pkg.Config)
	summary = fmt.Sprintf("%s\nPath: %s", summary, utils.ColoredString(pkg.Path, color.FgCyan))
	gui.renderString("secondary", summary)
	gui.activateContextView(pkg.ID())
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
		displayStrings := presentation.GetPackageListDisplayStrings(gui.State.Packages, gui.linkPathMap())
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

func (gui *Gui) handleCheckoutPackage() error {
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

func (gui *Gui) handleLinkPackage() error {
	// if it's the current package we should globally link it, otherwise we should link it to here
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	var cmdStr string
	if selectedPkg == gui.currentPackage() {
		return gui.surfaceError(errors.New("Cannot link a package to itself"))
	}

	if gui.linkPathMap()[selectedPkg.Path] {
		cmdStr = fmt.Sprintf("npm unlink --no-save %s", selectedPkg.Config.Name)
	} else {
		if !selectedPkg.LinkedGlobally {
			cmdStr = fmt.Sprintf("npm link %s", selectedPkg.Path)
		} else {
			cmdStr = fmt.Sprintf("npm link %s", selectedPkg.Config.Name)
		}
	}

	return gui.newMainCommand(cmdStr, selectedPkg.ID())
}

func (gui *Gui) handleGlobalLinkPackage() error {
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	if selectedPkg != gui.currentPackage() {
		return gui.surfaceError(errors.New("You can only globally link the current package. Hit space on this package to make it the current package."))
	}

	var cmdStr string
	if selectedPkg.LinkedGlobally {
		cmdStr = "npm unlink"
	} else {
		cmdStr = "npm link"
	}

	return gui.newMainCommand(cmdStr, selectedPkg.ID())
}

func (gui *Gui) handleInstall() error {
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	var cmdStr string
	if selectedPkg == gui.currentPackage() {
		cmdStr = "npm install"
	} else {
		cmdStr = "npm install --prefix " + selectedPkg.Path
	}

	return gui.newMainCommand(cmdStr, selectedPkg.ID())
}

func (gui *Gui) handleBuild() error {
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	var cmdStr string
	if selectedPkg == gui.currentPackage() {
		cmdStr = "npm run build"
	} else {
		cmdStr = "npm run build --prefix " + selectedPkg.Path
	}

	return gui.newMainCommand(cmdStr, selectedPkg.ID())
}

func (gui *Gui) handleOpenPackageConfig() error {
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	return gui.openFile(selectedPkg.ConfigPath())
}

func (gui *Gui) handleRemovePackage() error {
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	if selectedPkg == gui.currentPackage() {
		return gui.createErrorPanel("Cannot remove current package")
	}

	return gui.createConfirmationPanel(createConfirmationPanelOpts{
		returnToView:       gui.getPackagesView(),
		title:              "Remove package",
		prompt:             "Do you want to remove this package from the list? It won't actually be removed from the filesystem, but as far as lazynpm is concerned it'll be as good as dead. You won't have to worry about it no more.",
		returnFocusOnClose: true,
		handleConfirm: func() error {
			return gui.removePackage(selectedPkg.Path)
		},
	})
}

func (gui *Gui) handleAddPackage() error {
	return gui.createPromptPanel(gui.getPackagesView(), "Add package path to add", "", func(input string) error {
		configPath := input
		if !strings.HasSuffix(configPath, "package.json") {
			configPath = filepath.Join(configPath, "package.json")
		}
		if !commands.FileExists(configPath) {
			return gui.createErrorPanel(fmt.Sprintf("%s not found", configPath))
		}

		return gui.addPackage(strings.TrimSuffix(input, "package.json"))
	})
}

func (gui *Gui) handlePackPackage() error {
	selectedPkg := gui.getSelectedPackage()
	if selectedPkg == nil {
		return nil
	}

	cmdStr := "npm pack"
	if selectedPkg != gui.currentPackage() {
		cmdStr = fmt.Sprintf("npm pack %s", selectedPkg.Path)
	}

	return gui.newMainCommand(cmdStr, selectedPkg.ID())
}
