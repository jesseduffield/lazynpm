package presentation

import (
	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/theme"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

func GetPackageListDisplayStrings(packages []*commands.Package, deps []*commands.Dependency) [][]string {
	lines := make([][]string, len(packages))

	// we need to work out all the link paths from the deps
	linkPathMap := map[string]bool{}
	for _, dep := range deps {
		if dep.Linked() {
			linkPathMap[dep.LinkPath] = true
		}
	}

	for i := range packages {
		pkg := packages[i]
		lines[i] = getPackageDisplayStrings(pkg, linkPathMap[pkg.Path])
	}

	return lines
}

func getPackageDisplayStrings(p *commands.Package, linkedToCurrentPackage bool) []string {
	attr := theme.DefaultTextColor
	if p.LinkedGlobally {
		attr = color.FgYellow
	}
	line := utils.ColoredString(p.Config.Name, attr)
	linkedArg := ""
	if linkedToCurrentPackage {
		linkedArg = utils.ColoredString("(linked)", color.FgCyan)
	}
	return []string{line, linkedArg, utils.ColoredString(p.Path, color.FgBlue)}
}
