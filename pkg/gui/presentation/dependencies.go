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

func getDepDisplayStrings(d *commands.Dependency) []string {

	localVersionCol := ""
	if d.Linked() {
		localVersionCol = utils.ColoredString("linked: "+d.LinkPath, color.FgCyan)
	} else if d.PackageConfig != nil {
		localVersionCol = utils.ColoredString(d.PackageConfig.Version, color.FgYellow)
	}

	return []string{d.Name, utils.ColoredString(d.Version, color.FgMagenta), localVersionCol}
}
