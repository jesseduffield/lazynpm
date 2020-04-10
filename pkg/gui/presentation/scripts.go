package presentation

import (
	"github.com/jesseduffield/lazynpm/pkg/commands"
)

func GetScriptListDisplayStrings(scripts []*commands.Script) [][]string {
	lines := make([][]string, len(scripts))

	for i := range scripts {
		lines[i] = getScriptDisplayStrings(scripts[i])
	}

	return lines
}

func getScriptDisplayStrings(p *commands.Script) []string {
	return []string{p.Name, p.Command}
}
