package presentation

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/theme"
	"github.com/jesseduffield/lazynpm/pkg/utils"
	"github.com/jesseduffield/semver/v3"
)

func GetDependencyListDisplayStrings(dependencies []*commands.Dependency, commandMap commands.CommandViewMap) [][]string {
	lines := make([][]string, len(dependencies))

	for i := range dependencies {
		dep := dependencies[i]
		lines[i] = getDepDisplayStrings(dep, commandMap[dep.ID()])
	}

	return lines
}

func getDepDisplayStrings(d *commands.Dependency, commandView *commands.CommandView) []string {

	localVersionCol := ""
	if d.Linked() {
		localVersionCol = utils.ColoredString("linked: "+d.LinkPath, color.FgCyan)
	} else if d.PackageConfig != nil {
		status, ok := semverStatus(d.PackageConfig.Version, d.Version)
		if ok {
			localVersionCol = utils.ColoredString(d.PackageConfig.Version, color.FgGreen)
		} else {
			localVersionCol = utils.ColoredString(fmt.Sprintf("%s%s", d.PackageConfig.Version, statusMap()[status]), color.FgYellow)
		}
	} else {
		localVersionCol = utils.ColoredString("missing", color.FgRed)
	}

	kindColorMap := map[string]color.Attribute{
		"prod":     color.FgYellow,
		"dev":      color.FgGreen,
		"peer":     color.FgMagenta,
		"optional": theme.DefaultTextColor,
	}

	return []string{commandView.Status(), d.Name, utils.ColoredString(d.Kind, kindColorMap[d.Kind]), utils.ColoredString(d.Version, color.FgMagenta), localVersionCol}
}

func statusMap() map[int]string {
	return map[int]string{
		semver.BAD_AHEAD:  " (ahead)",
		semver.BAD_BEHIND: " (behind)",
		semver.BAD_EQUAL:  " (equal)",
		semver.BAD:        "",
	}
}

func semverStatus(version, constraint string) (int, bool) {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		// could have a formatted message here but too lazy
		return semver.BAD, false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		// could have a formatted message here but too lazy
		return semver.BAD, false
	}

	status := c.Status(v)

	return status, status == semver.GOOD
}
