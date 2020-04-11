package presentation

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/theme"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

func GetPackageListDisplayStrings(packages []*commands.Package, linkPathMap map[string]bool) [][]string {
	lines := make([][]string, len(packages))

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

func PackageSummary(pkgConfig commands.PackageConfig) string {
	output := ""
	if pkgConfig.Name != "" {
		output = fmt.Sprintf("Name: %s", utils.ColoredString(pkgConfig.Name, color.FgYellow))
	}
	if pkgConfig.Description != "" {
		output = fmt.Sprintf("%s\nDescription: %s", output, utils.ColoredString(pkgConfig.Description, color.FgCyan))
	}
	authorStr := pkgConfig.Author.ToString()
	if authorStr != "" {
		output = fmt.Sprintf("%s\nAuthor: %s", output, utils.ColoredString(authorStr, color.FgGreen))
	}
	repoStr := pkgConfig.Repository.ToString()
	if repoStr != "" {
		output = fmt.Sprintf("%s\nRepo: %s", output, utils.ColoredString(repoStr, color.FgRed))
	}
	if pkgConfig.Version != "" {
		output = fmt.Sprintf("%s\nVersion: %s", output, utils.ColoredString(pkgConfig.Version, color.FgYellow))
	}
	return output
}
