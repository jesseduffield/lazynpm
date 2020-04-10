package presentation

import (
	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

func GetDependencyListDisplayStrings(dependencies []*commands.Dependency) [][]string {
	lines := make([][]string, len(dependencies))

	for i := range dependencies {
		lines[i] = getDepDisplayStrings(dependencies[i])
	}

	return lines
}

func getDepDisplayStrings(p *commands.Dependency) []string {

	localVersionCol := utils.ColoredString(p.LocalVersion, color.FgYellow)
	if p.Linked() {
		localVersionCol = utils.ColoredString(p.LinkPath, color.FgCyan)
	}

	return []string{p.Name, utils.ColoredString(p.Version, color.FgMagenta), localVersionCol}
}
