package presentation

import (
	"strings"

	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/utils"
	"github.com/jesseduffield/semver"
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
		status, ok := semverStatus(d.PackageConfig.Version, d.Version)
		if ok {
			localVersionCol = utils.ColoredString(d.PackageConfig.Version, color.FgGreen)
		} else {
			localVersionCol = utils.ColoredString(d.PackageConfig.Version+" "+status, color.FgYellow)
		}
	} else {
		localVersionCol = utils.ColoredString("missing", color.FgRed)
	}

	return []string{d.Name, utils.ColoredString(d.Version, color.FgMagenta), localVersionCol}
}

func semverStatus(version, constraint string) (string, bool) {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return "error parsing constraint", false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return "error parsing version", false
	}

	ok, errors := c.Validate(v)
	if ok {
		return "", true
	}

	messages := make([]string, len(errors))
	for i, err := range errors {
		messages[i] = err.Error()
	}

	return strings.Join(messages, ","), false
}
