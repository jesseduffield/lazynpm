package gui

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
)

// list panel functions

func (gui *Gui) getSelectedDependency() *commands.Dependency {
	currentPackage := gui.currentPackage()

	deps := currentPackage.SortedDependencies()
	if len(deps) == 0 {
		return nil
	}
	return deps[gui.State.Panels.Deps.SelectedLine]
}

func (gui *Gui) handleDepSelect(g *gocui.Gui, v *gocui.View) error {
	dep := gui.getSelectedDependency()
	if dep == nil {
		gui.getMainView().Title = ""
		return gui.newStringTask("main", gui.Tr.SLocalize("NoDependencies"))
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
