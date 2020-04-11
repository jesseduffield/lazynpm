package gui

import (
	"fmt"

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
		gui.getMainView().Title = ""
		return gui.newStringTask("main", gui.Tr.SLocalize("NoDependencies"))
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
