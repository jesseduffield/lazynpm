package presentation

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/theme"
	"github.com/jesseduffield/lazynpm/pkg/utils"
	"github.com/jesseduffield/semver/v3"
)

func GetDependencyListDisplayStrings(dependencies []*commands.Dependency, commandMap commands.CommandViewMap, wide bool) [][]string {
	lines := make([][]string, len(dependencies))

	for i := range dependencies {
		dep := dependencies[i]
		lines[i] = getDepDisplayStrings(dep, commandMap[dep.ID()], wide)
	}

	return lines
}

func getDepDisplayStrings(d *commands.Dependency, commandView *commands.CommandView, wide bool) []string {
	localVersionCol := ""
	if d.Linked() {
		localVersionCol = utils.ColoredString("linked: "+d.LinkPath, color.FgCyan)
	} else if d.PackageConfig != nil {
		status, ok := semverStatus(d.PackageConfig.Version, d.Constraint)
		if ok {
			localVersionCol = utils.ColoredString(d.PackageConfig.Version, color.FgGreen)
		} else {
			localVersionCol = utils.ColoredString(fmt.Sprintf("%s%s", d.PackageConfig.Version, statusMap()[status]), color.FgYellow)
		}
	} else {
		localVersionCol = utils.ColoredString("missing", color.FgRed)
	}

	return []string{
		commandView.Status(),
		utils.ColoredString(truncateWithEllipsis(d.Name, 30, wide), KindColor(d.Kind)),
		utils.ColoredString(truncateWithEllipsis(d.Constraint, 20, wide), color.FgMagenta),
		localVersionCol,
	}
}

func truncateWithEllipsis(str string, limit int, wide bool) string {
	if wide {
		return str
	}

	return utils.TruncateWithEllipsis(str, limit)
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

func KindColor(kind string) color.Attribute {
	return map[string]color.Attribute{
		"prod":     theme.DefaultTextColor,
		"dev":      color.FgGreen,
		"optional": color.FgCyan,
		"peer":     color.FgMagenta,
	}[kind]
}
