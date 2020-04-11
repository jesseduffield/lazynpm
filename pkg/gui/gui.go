package gui

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"

	// "io"
	// "io/ioutil"

	"os/exec"
	"strings"
	"time"

	"github.com/go-errors/errors"

	// "strings"

	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/config"
	"github.com/jesseduffield/lazynpm/pkg/i18n"
	"github.com/jesseduffield/lazynpm/pkg/tasks"
	"github.com/jesseduffield/lazynpm/pkg/theme"
	"github.com/jesseduffield/lazynpm/pkg/updates"
	"github.com/jesseduffield/lazynpm/pkg/utils"
	"github.com/mattn/go-runewidth"
	"github.com/sirupsen/logrus"
)

const (
	SCREEN_NORMAL int = iota
	SCREEN_HALF
	SCREEN_FULL
)

const StartupPopupVersion = 1

// OverlappingEdges determines if panel edges overlap
var OverlappingEdges = false

// SentinelErrors are the errors that have special meaning and need to be checked
// by calling functions. The less of these, the better
type SentinelErrors struct {
	ErrSubProcess error
	ErrRestart    error
}

// GenerateSentinelErrors makes the sentinel errors for the gui. We're defining it here
// because we can't do package-scoped errors with localization, and also because
// it seems like package-scoped variables are bad in general
// https://dave.cheney.net/2017/06/11/go-without-package-scoped-variables
// In the future it would be good to implement some of the recommendations of
// that article. For now, if we don't need an error to be a sentinel, we will just
// define it inline. This has implications for error messages that pop up everywhere
// in that we'll be duplicating the default values. We may need to look at
// having a default localisation bundle defined, and just using keys-only when
// localising things in the code.
func (gui *Gui) GenerateSentinelErrors() {
	gui.Errors = SentinelErrors{
		ErrSubProcess: errors.New("running subprocess"),
		ErrRestart:    errors.New("restarting"),
	}
}

// Teml is short for template used to make the required map[string]interface{} shorter when using gui.Tr.SLocalize and gui.Tr.TemplateLocalize
type Teml i18n.Teml

// Gui wraps the gocui Gui object which handles rendering and events
type Gui struct {
	g                    *gocui.Gui
	Log                  *logrus.Entry
	NpmManager           *commands.NpmManager
	OSCommand            *commands.OSCommand
	SubProcess           *exec.Cmd
	State                *guiState
	Config               config.AppConfigurer
	Tr                   *i18n.Localizer
	Errors               SentinelErrors
	Updater              *updates.Updater
	statusManager        *statusManager
	waitForIntro         sync.WaitGroup
	viewBufferManagerMap map[string]*tasks.ViewBufferManager
	stopChan             chan struct{}
}

type packagesPanelState struct {
	SelectedLine int
}

type depsPanelState struct {
	SelectedLine int
}

type scriptsPanelState struct {
	SelectedLine int
}

type menuPanelState struct {
	SelectedLine int
	OnPress      func(g *gocui.Gui, v *gocui.View) error
}

type panelStates struct {
	Packages *packagesPanelState
	Deps     *depsPanelState
	Scripts  *scriptsPanelState
	Menu     *menuPanelState
}

type searchingState struct {
	view         *gocui.View
	isSearching  bool
	searchString string
}

type guiState struct {
	Packages          []*commands.Package
	Deps              []*commands.Dependency
	MenuItemCount     int // can't store the actual list because it's of interface{} type
	PreviousView      string
	Updating          bool
	Panels            *panelStates
	MainContext       string // used to keep the main and secondary views' contexts in sync
	RetainOriginalDir bool
	Searching         searchingState
	ScreenMode        int
	SideView          *gocui.View
	Ptmx              *os.File
	PrevMainWidth     int
	PrevMainHeight    int
	OldInformation    string
	StartupStage      int // one of INITIAL and COMPLETE. Allows us to not load everything at once
	CurrentPackageIdx int
}

func (gui *Gui) resetState() {
	gui.State = &guiState{
		Packages:     make([]*commands.Package, 0),
		PreviousView: "packages",
		Panels: &panelStates{
			Packages: &packagesPanelState{SelectedLine: 0},
			Deps:     &depsPanelState{SelectedLine: 0},
			Scripts:  &scriptsPanelState{SelectedLine: 0},
			Menu:     &menuPanelState{SelectedLine: 0},
		},
		SideView: nil,
		Ptmx:     nil,
	}
}

// for now the split view will always be on
// NewGui builds a new gui handler
func NewGui(log *logrus.Entry, gitCommand *commands.NpmManager, oSCommand *commands.OSCommand, tr *i18n.Localizer, config config.AppConfigurer, updater *updates.Updater) (*Gui, error) {
	gui := &Gui{
		Log:                  log,
		NpmManager:           gitCommand,
		OSCommand:            oSCommand,
		Config:               config,
		Tr:                   tr,
		Updater:              updater,
		statusManager:        &statusManager{},
		viewBufferManagerMap: map[string]*tasks.ViewBufferManager{},
	}

	gui.resetState()

	gui.GenerateSentinelErrors()

	return gui, nil
}

// Run setup the gui with keybindings and start the mainloop
func (gui *Gui) Run() error {
	gui.resetState()

	g, err := gocui.NewGui(gocui.Output256, OverlappingEdges, gui.Log)
	if err != nil {
		return err
	}
	defer g.Close()

	gui.State.ScreenMode = SCREEN_NORMAL
	g.OnSearchEscape = gui.onSearchEscape
	g.SearchEscapeKey = gui.getKey("universal.return")
	g.NextSearchMatchKey = gui.getKey("universal.nextMatch")
	g.PrevSearchMatchKey = gui.getKey("universal.prevMatch")

	gui.stopChan = make(chan struct{})

	g.ASCII = runtime.GOOS == "windows" && runewidth.IsEastAsian()

	if gui.Config.GetUserConfig().GetBool("gui.mouseEvents") {
		g.Mouse = true
	}

	gui.g = g // TODO: always use gui.g rather than passing g around everywhere

	if err := gui.setColorScheme(); err != nil {
		return err
	}

	popupTasks := []func(chan struct{}) error{}
	if gui.Config.GetUserConfig().GetString("reporting") == "undetermined" {
		popupTasks = append(popupTasks, gui.promptAnonymousReporting)
	}
	configPopupVersion := gui.Config.GetUserConfig().GetInt("StartupPopupVersion")
	// -1 means we've disabled these popups
	if configPopupVersion != -1 && configPopupVersion < StartupPopupVersion {
		popupTasks = append(popupTasks, gui.showShamelessSelfPromotionMessage)
	}
	gui.showInitialPopups(popupTasks)

	gui.waitForIntro.Add(1)

	gui.goEvery(time.Millisecond*250, gui.stopChan, gui.refreshPackages)
	gui.goEvery(time.Millisecond*50, gui.stopChan, gui.refreshScreen)

	g.SetManager(gocui.ManagerFunc(gui.layout), gocui.ManagerFunc(gui.getFocusLayout()))

	if err = gui.keybindings(g); err != nil {
		return err
	}

	gui.Log.Warn("starting main loop")

	err = g.MainLoop()
	return err
}

func (gui *Gui) refreshScreen() error {
	gui.g.Update(func(*gocui.Gui) error {
		return nil
	})
	return nil
}

// RunWithSubprocesses loops, instantiating a new gocui.Gui with each iteration
// if the error returned from a run is a ErrSubProcess, it runs the subprocess
// otherwise it handles the error, possibly by quitting the application
func (gui *Gui) RunWithSubprocesses() error {
	for {
		if err := gui.Run(); err != nil {
			for _, manager := range gui.viewBufferManagerMap {
				manager.Close()
			}
			gui.viewBufferManagerMap = map[string]*tasks.ViewBufferManager{}

			close(gui.stopChan)

			if err == gocui.ErrQuit {
				if !gui.State.RetainOriginalDir {
					if err := gui.recordCurrentDirectory(); err != nil {
						return err
					}
				}

				break
			} else if err == gui.Errors.ErrRestart {
				continue
			} else if err == gui.Errors.ErrSubProcess {
				if err := gui.runCommand(); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	return nil
}

func (gui *Gui) runCommand() error {
	gui.SubProcess.Stdout = os.Stdout
	gui.SubProcess.Stderr = os.Stdout
	gui.SubProcess.Stdin = os.Stdin

	fmt.Fprintf(os.Stdout, "\n%s\n\n", utils.ColoredString("+ "+strings.Join(gui.SubProcess.Args, " "), color.FgBlue))

	if err := gui.SubProcess.Run(); err != nil {
		// not handling the error explicitly because usually we're going to see it
		// in the output anyway
		gui.Log.Error(err)
	}

	gui.SubProcess.Stdout = ioutil.Discard
	gui.SubProcess.Stderr = ioutil.Discard
	gui.SubProcess.Stdin = nil
	gui.SubProcess = nil

	fmt.Fprintf(os.Stdout, "\n%s", utils.ColoredString(gui.Tr.SLocalize("pressEnterToReturn"), color.FgGreen))
	fmt.Scanln() // wait for enter press

	return nil
}

func (gui *Gui) loadNewRepo() error {
	gui.Updater.CheckForNewUpdate(gui.onBackgroundUpdateCheckFinish, false)
	if err := gui.updateRecentRepoList(); err != nil {
		return err
	}
	gui.waitForIntro.Done()

	if err := gui.refreshPackages(); err != nil {
		return err
	}

	return nil
}

// updateRecentRepoList registers the fact that we opened lazynpm in this package,
// so that appears in the packages view next time we open the program in another package
func (gui *Gui) updateRecentRepoList() error {
	ok, err := gui.NpmManager.ChdirToPackageRoot()
	if err != nil {
		return err
	}
	if ok {
		currentPackagePath, err := os.Getwd()
		if err != nil {
			return err
		}
		return gui.sendPackageToTop(currentPackagePath)
	}

	recentPackages := gui.Config.GetAppState().RecentPackages
	if len(recentPackages) > 0 {
		// TODO: ensure this actually contains a package.json file (meaning it won't be filtered out)
		return os.Chdir(recentPackages[0])
	}
	return errors.New("Must open lazynpm in an npm package")
}

func (gui *Gui) sendPackageToTop(path string) error {
	// in case we're not already there, chdir to path
	if err := os.Chdir(path); err != nil {
		return err
	}

	recentPackages := gui.Config.GetAppState().RecentPackages
	isNew, recentPackages := newRecentPackagesList(recentPackages, path)
	gui.Config.SetIsNewPackage(isNew)
	gui.Config.GetAppState().RecentPackages = recentPackages
	return gui.Config.SaveAppState()
}

// newRecentPackagesList returns a new repo list with a new entry but only when it doesn't exist yet
func newRecentPackagesList(recentPackages []string, currentPackage string) (bool, []string) {
	isNew := true
	newPackages := []string{currentPackage}
	for _, pkg := range recentPackages {
		if pkg != currentPackage {
			newPackages = append(newPackages, pkg)
		} else {
			isNew = false
		}
	}
	return isNew, newPackages
}

func (gui *Gui) showInitialPopups(tasks []func(chan struct{}) error) {
	gui.waitForIntro.Add(len(tasks))
	done := make(chan struct{})

	go func() {
		for _, task := range tasks {
			go func() {
				if err := task(done); err != nil {
					_ = gui.surfaceError(err)
				}
			}()

			<-done
			gui.waitForIntro.Done()
		}
	}()
}

func (gui *Gui) showShamelessSelfPromotionMessage(done chan struct{}) error {
	onConfirm := func(g *gocui.Gui, v *gocui.View) error {
		done <- struct{}{}
		return gui.Config.WriteToUserConfig("startupPopupVersion", StartupPopupVersion)
	}

	return gui.createConfirmationPanel(gui.g, nil, true, gui.Tr.SLocalize("ShamelessSelfPromotionTitle"), gui.Tr.SLocalize("ShamelessSelfPromotionMessage"), onConfirm, onConfirm)
}

func (gui *Gui) promptAnonymousReporting(done chan struct{}) error {
	return gui.createConfirmationPanel(gui.g, nil, true, gui.Tr.SLocalize("AnonymousReportingTitle"), gui.Tr.SLocalize("AnonymousReportingPrompt"), func(g *gocui.Gui, v *gocui.View) error {
		done <- struct{}{}
		return gui.Config.WriteToUserConfig("reporting", "on")
	}, func(g *gocui.Gui, v *gocui.View) error {
		done <- struct{}{}
		return gui.Config.WriteToUserConfig("reporting", "off")
	})
}

func (gui *Gui) goEvery(interval time.Duration, stop chan struct{}, function func() error) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = function()
			case <-stop:
				return
			}
		}
	}()
}

// setColorScheme sets the color scheme for the app based on the user config
func (gui *Gui) setColorScheme() error {
	userConfig := gui.Config.GetUserConfig()
	theme.UpdateTheme(userConfig)

	gui.g.FgColor = theme.InactiveBorderColor
	gui.g.SelFgColor = theme.ActiveBorderColor

	return nil
}
