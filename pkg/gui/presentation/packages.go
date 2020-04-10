package presentation

import (
	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/theme"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

func GetPackageListDisplayStrings(packages []*commands.Package) [][]string {
	lines := make([][]string, len(packages))

	for i := range packages {
		lines[i] = getPackageDisplayStrings(packages[i])
	}

	return lines
}

func getPackageDisplayStrings(p *commands.Package) []string {
	line := utils.ColoredString(p.Config.Name, theme.DefaultTextColor)
	if p.Linked {
		line += utils.ColoredString(" (linked)", color.FgCyan)
	}
	return []string{line}
}
