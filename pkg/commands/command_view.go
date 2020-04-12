package commands

import (
	"os/exec"

	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/utils"
)

type CommandView struct {
	// not super keen on having this dependency on gocui here but alas
	View      *gocui.View
	Cmd       *exec.Cmd
	Cancelled bool
}

func (cv *CommandView) Status() string {
	if cv == nil {
		return ""
	}

	if cv.Cancelled {
		return utils.ColoredString("*", color.FgRed, color.Bold)
	}

	if cv.Cmd.ProcessState == nil {
		return utils.ColoredString(utils.Loader(), color.FgCyan, color.Bold)
	} else {
		if cv.Cmd.ProcessState.Success() {
			return utils.ColoredString("!", color.FgGreen, color.Bold)
		} else {
			return utils.ColoredString("X", color.FgRed, color.Bold)
		}
	}
}

func (cv *CommandView) Running() bool {
	return cv.Cmd.ProcessState == nil
}

type CommandViewMap map[string]*CommandView
