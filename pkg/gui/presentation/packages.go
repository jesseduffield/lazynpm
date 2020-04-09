package presentation

import (
	"github.com/jesseduffield/lazynpm/pkg/commands"
)

func GetPackageListDisplayStrings(packages []*commands.Package) [][]string {
	lines := make([][]string, len(packages))

	for i := range packages {
		lines[i] = getFileDisplayStrings(packages[i])
	}

	return lines
}

func getFileDisplayStrings(p *commands.Package) []string {
	return []string{p.Name}
}
