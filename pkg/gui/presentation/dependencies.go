package presentation

import (
	"github.com/jesseduffield/lazynpm/pkg/commands"
)

func GetDependencyListDisplayStrings(dependencies []*commands.Dependency) [][]string {
	lines := make([][]string, len(dependencies))

	for i := range dependencies {
		lines[i] = getDepDisplayStrings(dependencies[i])
	}

	return lines
}

func getDepDisplayStrings(p *commands.Dependency) []string {
	return []string{p.Name, p.Version}
}
