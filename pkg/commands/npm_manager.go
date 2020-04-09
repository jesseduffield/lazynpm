package commands

import (
	"github.com/jesseduffield/lazynpm/pkg/config"
	"github.com/jesseduffield/lazynpm/pkg/i18n"
	"github.com/sirupsen/logrus"
)

// NpmManager is our main git interface
type NpmManager struct {
	Log       *logrus.Entry
	OSCommand *OSCommand
	Tr        *i18n.Localizer
	Config    config.AppConfigurer
}

// NewNpmManager it runs git commands
func NewNpmManager(log *logrus.Entry, osCommand *OSCommand, tr *i18n.Localizer, config config.AppConfigurer) (*NpmManager, error) {
	return &NpmManager{
		Log:       log,
		OSCommand: osCommand,
		Tr:        tr,
		Config:    config,
	}, nil
}

func (m *NpmManager) GetPackages() ([]*Package, error) {
	return nil, nil
}
