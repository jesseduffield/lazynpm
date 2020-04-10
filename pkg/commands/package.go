package commands

import (
	"encoding/json"
	"sort"
	"strings"
)

type PackageConfigInput struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	License     string   `json:"license"`
	Private     bool     `json:"private"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
	Keywords    []string `json:"keywords"`
	Os          []string `json:"os"`
	Cpu         []string `json:"cpu"`
	Main        string   `json:"main"`
	Engines     struct {
		Node string `json:"node"`
		Npm  string `json:"npm"`
	} `json:"engines"`
	Scripts         map[string]string `json:"scripts"`
	RawRepository   json.RawMessage   `json:"repository"`
	RawAuthor       json.RawMessage   `json:"author"`
	RawContributors []json.RawMessage `json:"contributors"`
	Bugs            struct {
		URL string `json:"url"`
	} `json:"bugs"`
	Deprecated           bool              `json:"deprecated"`
	Homepage             string            `json:"homepage"`
	Directories          map[string]string `json:"directories"`
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	BundledDependencies  []string          `json:"bundleDependencies"`
}

type Author struct {
	Name  string
	Email string
	Url   string
	// if a string rather than an object was given we'll store it in SingleLine
	SingleLine string
}

type Repository struct {
	Type string
	URL  string
	// if a string rather than an object was given we'll store it in SingleLine
	SingleLine string
}

type Package struct {
	Config PackageConfig
	Path   string
	// for when something is linked to the global node_modules folder
	LinkedGlobally bool
}

type PackageConfig struct {
	Name        string
	Version     string
	License     string
	Private     bool
	Description string
	Files       []string
	Keywords    []string
	Os          []string
	Cpu         []string
	Main        string
	Engines     struct {
		Node string
		Npm  string
	}
	Scripts      map[string]string
	Repository   Repository
	Author       Author
	Contributors []Author
	Bugs         struct {
		URL string
	}
	Deprecated           bool
	Homepage             string
	Directories          map[string]string
	Dependencies         map[string]string
	DevDependencies      map[string]string
	PeerDependencies     map[string]string
	OptionalDependencies map[string]string
	BundledDependencies  []string
}

func (p *Package) SortedDepsGeneric(depMap map[string]string) []*Dependency {
	deps := make([]*Dependency, 0, len(depMap))
	for name, version := range depMap {
		deps = append(deps, &Dependency{Name: name, Version: version})
	}
	sort.Slice(deps, func(i, j int) bool { return strings.Compare(deps[i].Name, deps[j].Name) < 0 })
	return deps
}

func (p *Package) SortedDependencies() []*Dependency {
	return p.SortedDepsGeneric(p.Config.Dependencies)
}
func (p *Package) SortedDevDependencies() []*Dependency {
	return p.SortedDepsGeneric(p.Config.DevDependencies)
}
func (p *Package) SortedPeerDependencies() []*Dependency {
	return p.SortedDepsGeneric(p.Config.PeerDependencies)
}
func (p *Package) SortedOptionalDependencies() []*Dependency {
	return p.SortedDepsGeneric(p.Config.OptionalDependencies)
}

func (p *Package) SortedScripts() []*Script {
	scripts := make([]*Script, 0, len(p.Config.Scripts))
	for name, command := range p.Config.Scripts {
		scripts = append(scripts, &Script{Name: name, Command: command})
	}
	sort.Slice(scripts, func(i, j int) bool { return strings.Compare(scripts[i].Name, scripts[j].Name) < 0 })
	return scripts
}
