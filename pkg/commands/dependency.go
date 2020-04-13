package commands

import (
	"fmt"
	"path/filepath"
)

type Dependency struct {
	Name              string
	Constraint        string
	LinkPath          string
	Present           bool
	PackageConfig     *PackageConfig
	Path              string
	Kind              string
	ParentPackagePath string
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

func KindKeyMap() map[string]string {
	return map[string]string{
		"prod":     "dependencies",
		"dev":      "devDependencies",
		"optional": "optionalDependencies",
		"peer":     "peerDependencies",
	}
}

func KindFlagMap() map[string]string {
	return map[string]string{
		"prod":     "--save-prod",
		"dev":      "--save-dev",
		"optional": "--save-optional",
	}
}

type KindFlag struct {
	Kind string
	Flag string
}

func KindFlags() []KindFlag {
	return []KindFlag{
		{Kind: "prod", Flag: "--save-prod"},
		{Kind: "dev", Flag: "--save-dev"},
		{Kind: "optional", Flag: "--save-optional"},
	}
}

func (d *Dependency) KindKey() string {
	return KindKeyMap()[d.Kind]
}
