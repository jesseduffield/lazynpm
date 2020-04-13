package presentation

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

func GetTarballListDisplayStrings(tarballs []*commands.Tarball, commandMap commands.CommandViewMap) [][]string {
	lines := make([][]string, len(tarballs))

	for i := range tarballs {
		tarball := tarballs[i]
		lines[i] = getTarballDisplayStrings(tarball, commandMap[tarball.ID()])
	}

	return lines
}

func getTarballDisplayStrings(t *commands.Tarball, commandView *commands.CommandView) []string {
	return []string{commandView.Status(), t.Name}
}

func TarballSummary(s *commands.Tarball) string {
	return fmt.Sprintf(
		"Name: %s\nPath: %s",
		utils.ColoredString(s.Name, color.FgYellow),
		utils.ColoredString(s.Path, color.FgCyan),
	)
}
