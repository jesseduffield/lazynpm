package gui

import (

	// "io"
	// "io/ioutil"

	// "strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/gui/presentation"
)

// list panel functions

func (gui *Gui) getSelectedFile() (*commands.File, error) {
	selectedLine := gui.State.Panels.Files.SelectedLine
	if selectedLine == -1 {
		return &commands.File{}, gui.Errors.ErrNoFiles
	}

	return gui.State.Files[selectedLine], nil
}

func (gui *Gui) selectFile(alreadySelected bool) error {
	gui.getFilesView().FocusPoint(0, gui.State.Panels.Files.SelectedLine)

	if gui.inDiffMode() {
		return gui.renderDiff()
	}

	file, err := gui.getSelectedFile()
	if err != nil {
		if err != gui.Errors.ErrNoFiles {
			return err
		}
		gui.State.SplitMainPanel = false
		gui.getMainView().Title = ""
		return gui.newStringTask("main", gui.Tr.SLocalize("NoChangedFiles"))
	}

	if !alreadySelected {
		if err := gui.resetOrigin(gui.getMainView()); err != nil {
			return err
		}
		if err := gui.resetOrigin(gui.getSecondaryView()); err != nil {
			return err
		}
	}

	if file.HasInlineMergeConflicts {
		gui.getMainView().Title = gui.Tr.SLocalize("MergeConflictsTitle")
		gui.State.SplitMainPanel = false
		return gui.refreshMergePanel()
	}

	if file.HasStagedChanges && file.HasUnstagedChanges {
		gui.State.SplitMainPanel = true
		gui.getMainView().Title = gui.Tr.SLocalize("UnstagedChanges")
		gui.getSecondaryView().Title = gui.Tr.SLocalize("StagedChanges")
		cmdStr := gui.GitCommand.DiffCmdStr(file, false, true)
		cmd := gui.OSCommand.ExecutableFromString(cmdStr)
		if err := gui.newPtyTask("secondary", cmd); err != nil {
			return err
		}
	} else {
		gui.State.SplitMainPanel = false
		if file.HasUnstagedChanges {
			gui.getMainView().Title = gui.Tr.SLocalize("UnstagedChanges")
		} else {
			gui.getMainView().Title = gui.Tr.SLocalize("StagedChanges")
		}
	}

	cmdStr := gui.GitCommand.DiffCmdStr(file, false, !file.HasUnstagedChanges && file.HasStagedChanges)
	cmd := gui.OSCommand.ExecutableFromString(cmdStr)
	if err := gui.newPtyTask("main", cmd); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) refreshFiles() error {
	selectedFile, _ := gui.getSelectedFile()

	filesView := gui.getFilesView()
	if filesView == nil {
		// if the filesView hasn't been instantiated yet we just return
		return nil
	}
	if err := gui.refreshStateFiles(); err != nil {
		return err
	}

	gui.g.Update(func(g *gocui.Gui) error {
		displayStrings := presentation.GetFileListDisplayStrings(gui.State.Files, gui.State.Diff.Ref)
		gui.renderDisplayStrings(filesView, displayStrings)

		if g.CurrentView() == filesView || (g.CurrentView() == gui.getMainView() && g.CurrentView().Context == "merging") {
			newSelectedFile, _ := gui.getSelectedFile()
			alreadySelected := newSelectedFile.Name == selectedFile.Name
			return gui.selectFile(alreadySelected)
		}
		return nil
	})

	return nil
}

// specific functions

// PrepareSubProcess - prepare a subprocess for execution and tell the gui to switch to it
func (gui *Gui) PrepareSubProcess(g *gocui.Gui, commands ...string) {
	gui.SubProcess = gui.GitCommand.PrepareCommitSubProcess()
	g.Update(func(g *gocui.Gui) error {
		return gui.Errors.ErrSubProcess
	})
}

func (gui *Gui) handleRefreshFiles(g *gocui.Gui, v *gocui.View) error {
	return gui.refreshSidePanels(refreshOptions{scope: []int{FILES}})
}

func (gui *Gui) refreshStateFiles() error {
	// get files to stage
	files := gui.GitCommand.GetStatusFiles()
	gui.State.Files = gui.GitCommand.MergeStatusFiles(gui.State.Files, files)

	gui.refreshSelectedLine(&gui.State.Panels.Files.SelectedLine, len(gui.State.Files))
	return nil
}

func (gui *Gui) onFilesPanelSearchSelect(selectedLine int) error {
	gui.State.Panels.Files.SelectedLine = selectedLine
	return gui.focusAndSelectFile(gui.g, gui.getFilesView())
}
