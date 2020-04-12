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
	if gui.State.CommandViewMap[viewName] == nil {
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
	gui.RefreshMutex.Lock()
	defer gui.RefreshMutex.Unlock()

	packagesView := gui.getPackagesView()
	if packagesView == nil {
		// if the filesView hasn't been instantiated yet we just return
		return nil
	}
	if err := gui.refreshStatePackages(); err != nil {
		return err
	}

	displayStrings := presentation.GetPackageListDisplayStrings(gui.State.Packages, gui.linkPathMap(), gui.State.CommandViewMap)
	gui.renderDisplayStrings(packagesView, displayStrings)

	displayStrings = presentation.GetDependencyListDisplayStrings(gui.State.Deps, gui.State.CommandViewMap)
	gui.renderDisplayStrings(gui.getDepsView(), displayStrings)

	displayStrings = presentation.GetScriptListDisplayStrings(gui.getScripts(), gui.State.CommandViewMap)
	gui.renderDisplayStrings(gui.getScriptsView(), displayStrings)

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
	gui.State.Packages, err = gui.NpmManager.GetPackages(gui.Config.GetAppState().RecentPackages, gui.State.Packages)
	if err != nil {
		return err
	}

	gui.State.Deps, err = gui.NpmManager.GetDeps(gui.currentPackage(), gui.State.Deps)
	if err != nil {
		return err
	}

	gui.refreshSelectedLine(&gui.State.Panels.Packages.SelectedLine, len(gui.State.Packages))
	return nil
}

func (gui *Gui) handleCheckoutPackage(pkg *commands.Package) error {
	if err := gui.sendPackageToTop(pkg.Path); err != nil {
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

func (gui *Gui) handleGlobalLinkPackage(pkg *commands.Package) error {
	if pkg != gui.currentPackage() {
		return gui.surfaceError(errors.New("You can only globally link the current package. Hit space on this package to make it the current package."))
	}

	var cmdStr string
	if pkg.LinkedGlobally {
		cmdStr = "npm unlink"
	} else {
		cmdStr = "npm link"
	}

	return gui.newMainCommand(cmdStr, pkg.ID())
}

func (gui *Gui) handleInstall(pkg *commands.Package) error {
	var cmdStr string
	if pkg == gui.currentPackage() {
		cmdStr = "npm install"
	} else {
		cmdStr = "npm install --prefix " + pkg.Path
	}

	return gui.newMainCommand(cmdStr, pkg.ID())
}

func (gui *Gui) handleBuild(pkg *commands.Package) error {
	var cmdStr string
	if pkg == gui.currentPackage() {
		cmdStr = "npm run build"
	} else {
		cmdStr = "npm run build --prefix " + pkg.Path
	}

	return gui.newMainCommand(cmdStr, pkg.ID())
}

func (gui *Gui) handleOpenPackageConfig(pkg *commands.Package) error {
	return gui.openFile(pkg.ConfigPath())
}

func (gui *Gui) handleRemovePackage(pkg *commands.Package) error {
	if pkg == gui.currentPackage() {
		return gui.createErrorPanel("Cannot remove current package")
	}

	return gui.createConfirmationPanel(createConfirmationPanelOpts{
		returnToView:       gui.getPackagesView(),
		title:              "Remove package",
		prompt:             "Do you want to remove this package from the list? It won't actually be removed from the filesystem, but as far as lazynpm is concerned it'll be as good as dead. You won't have to worry about it no more.",
		returnFocusOnClose: true,
		handleConfirm: func() error {
			return gui.finalStep(gui.removePackage(pkg.Path))
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

		return gui.finalStep(gui.addPackage(strings.TrimSuffix(input, "package.json")))
	})
}

func (gui *Gui) handlePackPackage(pkg *commands.Package) error {
	cmdStr := "npm pack"
	if pkg != gui.currentPackage() {
		cmdStr = fmt.Sprintf("npm pack %s", pkg.Path)
	}

	return gui.newMainCommand(cmdStr, pkg.ID())
}

func (gui *Gui) selectedPackageID() string {
	pkg := gui.getSelectedPackage()
	if pkg == nil {
		return ""
	}

	return pkg.ID()
}

func (gui *Gui) handlePublishPackage(pkg *commands.Package) error {
	cmdStr := "npm publish"

	tagPrompt := func() error {
		return gui.createPromptPanel(gui.getPackagesView(), "Enter tag name (leave blank for no tag)", "", func(tag string) error {
			if tag != "" {
				cmdStr = fmt.Sprintf("%s --tag=%s", cmdStr, tag)
			}
			cmdStr = fmt.Sprintf("%s %s", cmdStr, pkg.Config.Name)
			return gui.newMainCommand(cmdStr, pkg.ID())
		})
	}

	if pkg.Scoped() {
		menuItems := []*menuItem{
			{
				displayStrings: []string{"restricted (default)", utils.ColoredString("--access=restricted", color.FgYellow)},
				onPress: func() error {
					cmdStr = fmt.Sprintf("%s --access=restricted", cmdStr)
					return tagPrompt()
				},
			},
			{
				displayStrings: []string{"public", utils.ColoredString("--access=public", color.FgYellow)},
				onPress: func() error {
					cmdStr = fmt.Sprintf("%s --access=public", cmdStr)
					return tagPrompt()
				},
			},
		}

		return gui.createMenu("Set access for publishing scoped package (npm publish)", menuItems, createMenuOptions{showCancel: true})
	}

	return tagPrompt()
}

func (gui *Gui) wrappedPackageHandler(f func(*commands.Package) error) func(*gocui.Gui, *gocui.View) error {
	return gui.wrappedHandler(func() error {
		pkg := gui.getSelectedPackage()
		if pkg == nil {
			return nil
		}

		return gui.finalStep(f(pkg))
	})
}
