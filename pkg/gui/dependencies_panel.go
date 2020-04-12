package gui

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/go-errors/errors"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/gui/presentation"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

// list panel functions

func (gui *Gui) getSelectedDependency() *commands.Dependency {
	if len(gui.State.Deps) == 0 {
		return nil
	}
	return gui.State.Deps[gui.State.Panels.Deps.SelectedLine]
}

func (gui *Gui) handleDepSelect(g *gocui.Gui, v *gocui.View) error {
	dep := gui.getSelectedDependency()
	if dep == nil {
		gui.printToMain(gui.Tr.SLocalize("NoDependencies"))
		return nil
	}
	if dep.PackageConfig != nil {
		summary := presentation.PackageSummary(*dep.PackageConfig)
		if dep.Linked() {
			summary = fmt.Sprintf("%s\nLinked to: %s", summary, utils.ColoredString(dep.LinkPath, color.FgCyan))
		}
		gui.renderString("secondary", summary)
	} else {
		gui.renderString("secondary", "dependency not present in node_modules")
	}
	gui.activateContextView(gui.depContextKey(dep))
	return nil
}

// linkPathMap returns the set of link paths of the current package's dependencies
func (gui *Gui) linkPathMap() map[string]bool {
	linkPathMap := map[string]bool{}
	for _, dep := range gui.State.Deps {
		if dep.Linked() {
			linkPathMap[dep.LinkPath] = true
		}
	}
	return linkPathMap
}

func (gui *Gui) handleDepInstall() error {
	dep := gui.getSelectedDependency()
	if dep == nil {
		return nil
	}

	cmdStr := fmt.Sprintf("npm install %s", dep.Name)
	return gui.newMainCommand(cmdStr, gui.depContextKey(dep))
}

func (gui *Gui) handleDepUpdate() error {
	dep := gui.getSelectedDependency()
	if dep == nil {
		return nil
	}

	cmdStr := fmt.Sprintf("npm update %s", dep.Name)
	return gui.newMainCommand(cmdStr, gui.depContextKey(dep))
}

func (gui *Gui) handleOpenDepPackageConfig() error {
	selectedDep := gui.getSelectedDependency()
	if selectedDep == nil {
		return nil
	}

	if selectedDep.PackageConfig == nil {
		return gui.surfaceError(errors.New("dependency not in node_modules"))
	}

	return gui.openFile(selectedDep.ConfigPath())
}

func (gui *Gui) handleDepUninstall() error {
	selectedDep := gui.getSelectedDependency()
	if selectedDep == nil {
		return nil
	}

	var menuItems []*menuItem

	if selectedDep.Kind == "peer" {
		// I have no idea how peer dependencies work, so we're just using the one option here
		uninstallStr := fmt.Sprintf("npm uninstall %s", selectedDep.Name)

		menuItems = []*menuItem{
			{
				displayStrings: []string{"uninstall", utils.ColoredString(uninstallStr, color.FgYellow)},
				onPress: func() error {
					return gui.newMainCommand(uninstallStr, gui.depContextKey(selectedDep))
				},
			},
		}
	} else {
		kindMap := map[string]string{
			"prod":     " --save",
			"dev":      " --save-dev",
			"optional": " --save-optional",
		}

		uninstallCmdStr := fmt.Sprintf("npm uninstall --no-save %s", selectedDep.Name)
		uninstallAndSaveCmdStr := fmt.Sprintf("npm uninstall%s %s", kindMap[selectedDep.Kind], selectedDep.Name)

		menuItems = []*menuItem{
			{
				displayStrings: []string{"uninstall and save", utils.ColoredString(uninstallAndSaveCmdStr, color.FgYellow)},
				onPress: func() error {
					return gui.newMainCommand(uninstallAndSaveCmdStr, gui.depContextKey(selectedDep))
				},
			},
			{
				displayStrings: []string{"just uninstall", utils.ColoredString(uninstallCmdStr, color.FgYellow)},
				onPress: func() error {
					return gui.newMainCommand(uninstallCmdStr, gui.depContextKey(selectedDep))
				},
			},
		}
	}

	return gui.createMenu("Uninstall dependency", menuItems, createMenuOptions{showCancel: true})
}

func (gui *Gui) depContextKey(dep *commands.Dependency) string {
	return fmt.Sprintf("package:%s|dep:%s|kind:%s", gui.currentPackage().Path, dep.Name, dep.Kind)
}
