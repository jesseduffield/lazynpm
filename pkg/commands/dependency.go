package commands

import (
	"fmt"
	"path/filepath"
)

type Dependency struct {
	Name          string
	Version       string
	LinkPath      string
	Present       bool
	PackageConfig *PackageConfig
	Path          string
	Kind          string
}

func (d *Dependency) Linked() bool {
	return d.LinkPath != ""
}

func (d *Dependency) ConfigPath() string {
	return filepath.Join(d.Path, "package.json")
}

func (d *Dependency) ID() string {
	return fmt.Sprintf("dep:%s|kind:%s", d.Path, d.Kind)
}
