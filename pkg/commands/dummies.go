package commands

import (
	"io/ioutil"

	"github.com/jesseduffield/lazynpm/pkg/config"
	"github.com/jesseduffield/lazynpm/pkg/i18n"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// This file exports dummy constructors for use by tests in other packages

// NewDummyOSCommand creates a new dummy OSCommand for testing
func NewDummyOSCommand() *OSCommand {
	return NewOSCommand(NewDummyLog(), NewDummyAppConfig())
}

// NewDummyAppConfig creates a new dummy AppConfig for testing
func NewDummyAppConfig() *config.AppConfig {
	appConfig := &config.AppConfig{
		Name:        "lazynpm",
		Version:     "unversioned",
		Commit:      "",
		BuildDate:   "",
		Debug:       false,
		BuildSource: "",
		UserConfig:  viper.New(),
	}
	_ = yaml.Unmarshal([]byte{}, appConfig.AppState)
	return appConfig
}

// NewDummyLog creates a new dummy Log for testing
func NewDummyLog() *logrus.Entry {
	log := logrus.New()
	log.Out = ioutil.Discard
	return log.WithField("test", "test")
}

// NewDummyNpmManager creates a new dummy NpmManager for testing
func NewDummyNpmManager() *NpmManager {
	return NewDummyNpmManagerWithOSCommand(NewDummyOSCommand())
}

// NewDummyNpmManagerWithOSCommand creates a new dummy NpmManager for testing
func NewDummyNpmManagerWithOSCommand(osCommand *OSCommand) *NpmManager {
	return &NpmManager{
		Log:       NewDummyLog(),
		OSCommand: osCommand,
		Tr:        i18n.NewLocalizer(NewDummyLog()),
		Config:    NewDummyAppConfig(),
	}
}
