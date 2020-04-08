package app

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jesseduffield/lazynpm/pkg/commands"
	"github.com/jesseduffield/lazynpm/pkg/config"
	"github.com/jesseduffield/lazynpm/pkg/gui"
	"github.com/jesseduffield/lazynpm/pkg/i18n"
	"github.com/jesseduffield/lazynpm/pkg/updates"
	"github.com/shibukawa/configdir"
	"github.com/sirupsen/logrus"
)

// App struct
type App struct {
	closers []io.Closer

	Config     config.AppConfigurer
	Log        *logrus.Entry
	OSCommand  *commands.OSCommand
	GitCommand *commands.GitCommand
	Gui        *gui.Gui
	Tr         *i18n.Localizer
	Updater    *updates.Updater // may only need this on the Gui
}

type errorMapping struct {
	originalError string
	newError      string
}

func newProductionLogger(config config.AppConfigurer) *logrus.Logger {
	log := logrus.New()
	log.Out = ioutil.Discard
	log.SetLevel(logrus.ErrorLevel)
	return log
}

func globalConfigDir() string {
	configDirs := configdir.New("jesseduffield", "lazynpm")
	configDir := configDirs.QueryFolders(configdir.Global)[0]
	return configDir.Path
}

func getLogLevel() logrus.Level {
	strLevel := os.Getenv("LOG_LEVEL")
	level, err := logrus.ParseLevel(strLevel)
	if err != nil {
		return logrus.DebugLevel
	}
	return level
}

func newDevelopmentLogger(config config.AppConfigurer) *logrus.Logger {
	log := logrus.New()
	log.SetLevel(getLogLevel())
	file, err := os.OpenFile(filepath.Join(globalConfigDir(), "development.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("unable to log to file") // TODO: don't panic (also, remove this call to the `panic` function)
	}
	log.SetOutput(file)
	return log
}

func newLogger(config config.AppConfigurer) *logrus.Entry {
	var log *logrus.Logger
	if config.GetDebug() || os.Getenv("DEBUG") == "TRUE" {
		log = newDevelopmentLogger(config)
	} else {
		log = newProductionLogger(config)
	}

	// highly recommended: tail -f development.log | humanlog
	// https://github.com/aybabtme/humanlog
	log.Formatter = &logrus.JSONFormatter{}

	return log.WithFields(logrus.Fields{
		"debug":     config.GetDebug(),
		"version":   config.GetVersion(),
		"commit":    config.GetCommit(),
		"buildDate": config.GetBuildDate(),
	})
}

// NewApp bootstrap a new application
func NewApp(config config.AppConfigurer) (*App, error) {
	app := &App{
		closers: []io.Closer{},
		Config:  config,
	}
	var err error
	app.Log = newLogger(config)
	app.Tr = i18n.NewLocalizer(app.Log)

	app.OSCommand = commands.NewOSCommand(app.Log, config)

	app.Updater, err = updates.NewUpdater(app.Log, config, app.OSCommand, app.Tr)
	if err != nil {
		return app, err
	}

	if err := app.setupPackage(); err != nil {
		return app, err
	}

	app.GitCommand, err = commands.NewGitCommand(app.Log, app.OSCommand, app.Tr, app.Config)
	if err != nil {
		return app, err
	}
	app.Gui, err = gui.NewGui(app.Log, app.GitCommand, app.OSCommand, app.Tr, config, app.Updater)
	if err != nil {
		return app, err
	}
	return app, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (app *App) setupPackage() error {
	// ensure we have a package.json file here or in an ancestor directory
	// if there's not, pick the first one from state or return an error
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if fileExists("package.json") {
			return nil
		}

		if err := os.Chdir(".."); err != nil {
			return err
		}

		newDir, err := os.Getwd()
		if err != nil {
			return err
		}
		if newDir == dir {
			return errors.New("must start lazynpm in an npm package")
		}
		dir = newDir
	}
}

func (app *App) Run() error {
	err := app.Gui.RunWithSubprocesses()
	return err
}

// Close closes any resources
func (app *App) Close() error {
	for _, closer := range app.closers {
		err := closer.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// KnownError takes an error and tells us whether it's an error that we know about where we can print a nicely formatted version of it rather than panicking with a stack trace
func (app *App) KnownError(err error) (string, bool) {
	errorMessage := err.Error()

	mappings := []errorMapping{}

	for _, mapping := range mappings {
		if strings.Contains(errorMessage, mapping.originalError) {
			return mapping.newError, true
		}
	}
	return "", false
}
