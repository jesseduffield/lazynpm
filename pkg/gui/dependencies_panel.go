package gui

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
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
		summary = fmt.Sprintf("%s\nConstraint: %s", summary, utils.ColoredString(dep.Constraint, color.FgMagenta))
		summary = fmt.Sprintf("%s\nType: %s", summary, utils.ColoredString(dep.KindKey(), presentation.KindColor(dep.Kind)))
		if dep.Linked() {
			summary = fmt.Sprintf("%s\nLinked to: %s", summary, utils.ColoredString(dep.LinkPath, color.FgCyan))
		}
		gui.renderString("secondary", summary)
	} else {
		gui.renderString("secondary", "dependency not present in node_modules")
	}
	gui.activateContextView(dep.ID())
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

func (gui *Gui) handleDepInstall(dep *commands.Dependency) error {
	cmdStr := fmt.Sprintf("npm install %s", dep.Name)
	return gui.newMainCommand(cmdStr, dep.ID())
}

func (gui *Gui) handleDepUpdate(dep *commands.Dependency) error {
	cmdStr := fmt.Sprintf("npm update %s", dep.Name)
	return gui.newMainCommand(cmdStr, dep.ID())
}

func (gui *Gui) handleOpenDepPackageConfig(dep *commands.Dependency) error {
	if dep.PackageConfig == nil {
		return gui.createErrorPanel("dependency not in node_modules")
	}

	return gui.openFile(dep.ConfigPath())
}

func (gui *Gui) handleDepUninstall(dep *commands.Dependency) error {
	var menuItems []*menuItem

	if dep.Kind == "peer" {
		// I have no idea how peer dependencies work, so we're just using the one option here
		uninstallStr := fmt.Sprintf("npm uninstall %s", dep.Name)

		menuItems = []*menuItem{
			{
				displayStrings: []string{"uninstall", utils.ColoredString(uninstallStr, color.FgYellow)},
				onPress: func() error {
					return gui.newMainCommand(uninstallStr, dep.ID())
				},
			},
		}
	} else {
		kindMap := map[string]string{
			"prod":     " --save",
			"dev":      " --save-dev",
			"optional": " --save-optional",
		}

		uninstallCmdStr := fmt.Sprintf("npm uninstall --no-save %s", dep.Name)
		uninstallAndSaveCmdStr := fmt.Sprintf("npm uninstall%s %s", kindMap[dep.Kind], dep.Name)

		menuItems = []*menuItem{
			{
				displayStrings: []string{"uninstall and save", utils.ColoredString(uninstallAndSaveCmdStr, color.FgYellow)},
				onPress: func() error {
					return gui.newMainCommand(uninstallAndSaveCmdStr, dep.ID())
				},
			},
			{
				displayStrings: []string{"just uninstall", utils.ColoredString(uninstallCmdStr, color.FgYellow)},
				onPress: func() error {
					return gui.newMainCommand(uninstallCmdStr, dep.ID())
				},
			},
		}
	}

	return gui.createMenu("Uninstall dependency", menuItems, createMenuOptions{showCancel: true})
}

func (gui *Gui) selectedDepID() string {
	selectedDep := gui.getSelectedDependency()
	if selectedDep == nil {
		return ""
	}

	return selectedDep.ID()
}

func (gui *Gui) wrappedDependencyHandler(f func(*commands.Dependency) error) func(*gocui.Gui, *gocui.View) error {
	return gui.wrappedHandler(func() error {
		dep := gui.getSelectedDependency()
		if dep == nil {
			return nil
		}

		return gui.finalStep(f(dep))
	})
}

func (gui *Gui) handleChangeDepType(dep *commands.Dependency) error {
	kindKeyMap := commands.KindKeyMap()
	kindFlags := commands.KindFlags()
	menuItems := make([]*menuItem, 0, len(kindFlags))
	for _, kindFlag := range kindFlags {
		kindFlag := kindFlag
		cmdStr := fmt.Sprintf("npm install %s %s", kindFlag.Flag, dep.Name)
		menuItems = append(menuItems, &menuItem{
			displayStrings: []string{kindKeyMap[kindFlag.Kind], utils.ColoredString(cmdStr, color.FgYellow)},
			onPress: func() error {
				return gui.newMainCommand(cmdStr, dep.ID())
			},
		})
	}

	return gui.createMenu("Change dependency type", menuItems, createMenuOptions{showCancel: true})
}

// this is admittedly a little weird. We're going to store the command against
// the dep where you initiated the command, but it has nothing to do with that dep.
func (gui *Gui) handleAddDependency(dep *commands.Dependency) error {
	prompt := func(cmdStr string) error {
		return gui.createPromptPanel(gui.getDepsView(), "enter dependency name", "", func(input string) error {
			newCmdStr := fmt.Sprintf("%s %s", cmdStr, input)
			return gui.newMainCommand(newCmdStr, dep.ID())
		})
	}

	kindKeyMap := commands.KindKeyMap()
	kindFlags := commands.KindFlags()
	menuItems := make([]*menuItem, 0, len(kindFlags))
	for _, kindFlag := range kindFlags {
		kindFlag := kindFlag
		cmdStr := fmt.Sprintf("npm install %s", kindFlag.Flag)
		menuItems = append(menuItems, &menuItem{
			displayStrings: []string{kindKeyMap[kindFlag.Kind], utils.ColoredString(cmdStr, color.FgYellow)},
			onPress: func() error {
				return prompt(cmdStr)
			},
		})
	}

	return gui.createMenu("Install dependency to:", menuItems, createMenuOptions{showCancel: true})
}

func (gui *Gui) handleEditDepConstraint(dep *commands.Dependency) error {
	return gui.createPromptPanel(gui.getDepsView(), "Edit constraint", dep.Constraint, func(input string) error {

		packageConfigPath := filepath.Join(dep.ParentPackagePath, "package.json")
		return gui.finalStep(gui.NpmManager.EditDepConstraint(dep, packageConfigPath, input))
	})
}
