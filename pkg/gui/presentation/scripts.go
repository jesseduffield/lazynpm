package presentation

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

func GetScriptListDisplayStrings(scripts []*commands.Script, commandMap commands.CommandViewMap) [][]string {
	lines := make([][]string, len(scripts))

	for i := range scripts {
		script := scripts[i]
		lines[i] = getScriptDisplayStrings(script, commandMap[script.ID()])
	}

	return lines
}

func getScriptDisplayStrings(p *commands.Script, commandView *commands.CommandView) []string {
	return []string{commandView.Status(), p.Name, utils.ColoredString(p.Command, color.FgBlue)}
}

func ScriptSummary(s *commands.Script) string {
	return fmt.Sprintf(
		"Name: %s\nCommand: %s",
		utils.ColoredString(s.Name, color.FgYellow),
		utils.ColoredString(s.Command, color.FgCyan),
	)
}
