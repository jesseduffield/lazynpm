package commands

import (
	"os/exec"

	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

type CommandView struct {
	// not super keen on having this dependency on gocui here but alas
	View *gocui.View
	Cmd  *exec.Cmd
}

func (cv *CommandView) Status() string {
	if cv == nil {
		return ""
	}

	if cv.Cmd.ProcessState == nil {
		return utils.ColoredString(utils.Loader(), color.FgCyan)
	} else {
		if cv.Cmd.ProcessState.Success() {
			return utils.ColoredString("!", color.FgGreen)
		} else {
			return utils.ColoredString("X", color.FgRed)
		}
	}
}
