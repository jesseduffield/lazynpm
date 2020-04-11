package commands

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/buger/jsonparser"
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
	NpmRoot   string
}

// NewNpmManager it runs git commands
func NewNpmManager(log *logrus.Entry, osCommand *OSCommand, tr *i18n.Localizer, config config.AppConfigurer) (*NpmManager, error) {
	output, err := osCommand.RunCommandWithOutput("npm root -g")
	if err != nil {
		return nil, err
	}
	npmRoot := strings.TrimSpace(output)

	return &NpmManager{
		Log:       log,
		OSCommand: osCommand,
		Tr:        tr,
		Config:    config,
		NpmRoot:   npmRoot,
	}, nil
}

func (m *NpmManager) IsLinked(name string, path string) (bool, error) {
	globalPath := filepath.Join(m.NpmRoot, name)
	fileInfo, err := os.Lstat(globalPath)
	if err != nil {
		if err == os.ErrNotExist {
			return false, nil
		}
		// swallowing error. For some reason we're getting 'no such file or directory' here despite checking for os.ErrNotExist
		return false, nil
	}

	isSymlink := fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink
	if isSymlink {
		linkedPath, err := os.Readlink(globalPath)
		if err != nil {
			return false, err
		}
		if linkedPath == path {
			return true, nil
		}
	}
	return false, nil
}

func (m *NpmManager) GetPackages(paths []string) ([]*Package, error) {

	pkgs := make([]*Package, 0, len(paths))

	for _, path := range paths {
		packageConfigPath := filepath.Join(path, "package.json")
		if !FileExists(packageConfigPath) {
			continue
		}

		file, err := os.OpenFile(packageConfigPath, os.O_RDONLY, 0644)
		if err != nil {
			m.Log.Error(err)
			continue
		}
		pkgConfig, err := UnmarshalPackageConfig(file)
		if err != nil {
			return nil, err
		}
		linked, err := m.IsLinked(pkgConfig.Name, path)
		if err != nil {
			return nil, err
		}

		pkgs = append(pkgs, &Package{
			Config:         *pkgConfig,
			Path:           path,
			LinkedGlobally: linked,
		})
	}
	return pkgs, nil
}

func (m *NpmManager) ChdirToPackageRoot() (bool, error) {
	dir, err := os.Getwd()
	if err != nil {
		return false, err
	}
	for {
		if FileExists("package.json") {
			return true, nil
		}

		if err := os.Chdir(".."); err != nil {
			return false, err
		}

		newDir, err := os.Getwd()
		if err != nil {
			return false, err
		}
		if newDir == dir {
			return false, nil
		}
		dir = newDir
	}
}

func (m *NpmManager) GetDeps(currentPkg *Package) ([]*Dependency, error) {
	deps := currentPkg.SortedDependencies()

	for _, dep := range deps {
		depPath := filepath.Join(currentPkg.Path, "node_modules", dep.Name)
		dep.Path = depPath
		fileInfo, err := os.Lstat(depPath)
		if err != nil {
			// must not be present in node modules
			m.Log.Error(err)
			continue
		}
		dep.Present = true

		// get the actual version of the package in node modules
		packageConfigPath := filepath.Join(depPath, "package.json")
		file, err := os.OpenFile(packageConfigPath, os.O_RDONLY, 0644)
		if err != nil {
			m.Log.Error(err)
			continue
		}
		pkgConfig, err := UnmarshalPackageConfig(file)
		if err != nil {
			// swallowing error
			m.Log.Error(err)
		} else {
			dep.PackageConfig = pkgConfig
		}

		isSymlink := fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink
		if !isSymlink {
			continue
		}

		linkPath, err := filepath.EvalSymlinks(depPath)
		if err != nil {
			return nil, err
		}
		dep.LinkPath = linkPath
	}

	return deps, nil
}

func (m *NpmManager) RemoveScript(scriptName string, packageJsonPath string) error {
	config, err := ioutil.ReadFile(packageJsonPath)
	if err != nil {
		return err
	}

	updatedConfig := jsonparser.Delete(config, "scripts", scriptName)

	return ioutil.WriteFile(packageJsonPath, updatedConfig, 0644)
}
