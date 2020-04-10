package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/jesseduffield/lazynpm/pkg/config"
	"github.com/jesseduffield/lazynpm/pkg/i18n"
	"github.com/jinzhu/copier"
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

func (m *NpmManager) UnmarshalPackage(r io.Reader) (*Package, error) {
	var pkgInput *PackageInput
	d := json.NewDecoder(r)
	if err := d.Decode(&pkgInput); err != nil {
		return nil, err
	}

	var pkg Package
	if err := copier.Copy(&pkg, &pkgInput); err != nil {
		return nil, err
	}

	isObject := func(b []byte) bool {
		return bytes.HasPrefix(b, []byte{'{'})
	}

	if isObject(pkgInput.RawAuthor) {
		err := json.Unmarshal(pkgInput.RawAuthor, &pkg.Author)
		if err != nil {
			return nil, err
		}
	} else {
		pkg.Author.SingleLine = string(pkgInput.RawAuthor)
	}

	for _, rawContributor := range pkgInput.RawContributors {
		var contributor *Author
		if isObject(rawContributor) {
			err := json.Unmarshal(rawContributor, contributor)
			if err != nil {
				return nil, err
			}
		} else {
			contributor = &Author{SingleLine: string(rawContributor)}
		}
		pkg.Contributors = append(pkg.Contributors, *contributor)
	}

	if isObject(pkgInput.RawRepository) {
		err := json.Unmarshal(pkgInput.RawRepository, &pkg.Repository)
		if err != nil {
			return nil, err
		}
	} else {
		pkg.Repository.SingleLine = string(pkgInput.RawRepository)
	}
	return &pkg, nil
}

func (m *NpmManager) GetPackages(paths []string) ([]*Package, error) {
	pkgs := make([]*Package, 0, len(paths))
	for _, path := range paths {
		packageJsonPath := filepath.Join(path, "package.json")
		if !FileExists(packageJsonPath) {
			m.Log.Error("package.json does not exist at " + packageJsonPath)
			continue
		}

		file, err := os.OpenFile(packageJsonPath, os.O_RDONLY, 0644)
		if err != nil {
			m.Log.Error(err)
			continue
		}
		pkg, err := m.UnmarshalPackage(file)
		if err != nil {
			return nil, err
		}

		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}
