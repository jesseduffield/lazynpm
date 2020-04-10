package presentation

import (
	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

func GetScriptListDisplayStrings(scripts []*commands.Script) [][]string {
	lines := make([][]string, len(scripts))

	for i := range scripts {
		lines[i] = getScriptDisplayStrings(scripts[i])
	}

	return lines
}

func getScriptDisplayStrings(p *commands.Script) []string {
	return []string{p.Name, utils.ColoredString(p.Command, color.FgBlue)}
}
