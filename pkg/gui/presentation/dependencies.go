package presentation

import (
	"github.com/Masterminds/semver"
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
		attr := color.FgYellow
		if semverGood(d.PackageConfig.Version, d.Version) {
			attr = color.FgGreen
		}
		localVersionCol = utils.ColoredString(d.PackageConfig.Version, attr)
	} else {
		localVersionCol = utils.ColoredString("missing", color.FgRed)
	}

	return []string{d.Name, utils.ColoredString(d.Version, color.FgMagenta), localVersionCol}
}

func semverGood(version, constraint string) bool {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}

	a := c.Check(v)

	return a
}
