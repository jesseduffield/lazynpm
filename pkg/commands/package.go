package commands

import (
	"encoding/json"
	"fmt"
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
	Scripts              map[string]string `json:"scripts"`
	RawRepository        json.RawMessage   `json:"repository"`
	RawAuthor            json.RawMessage   `json:"author"`
	RawContributors      []json.RawMessage `json:"contributors"`
	RawBugs              json.RawMessage   `json:"bugs"`
	Deprecated           bool              `json:"deprecated"`
	Homepage             string            `json:"homepage"`
	Directories          map[string]string `json:"directories"`
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	BundledDependencies  []string          `json:"bundleDependencies"`
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
		Url string `json:"url"`
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

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Url   string `json:"url"`
	// if a string rather than an object was given we'll store it in SingleLine
	SingleLine string
}

func (a Author) ToString() string {
	if a.SingleLine != "" {
		return a.SingleLine
	}
	output := a.Name
	if a.Email != "" {
		output = fmt.Sprintf("%s <%s>", output, a.Email)
	}
	if a.Url != "" {
		output = fmt.Sprintf("%s (%s)", output, a.Url)
	}
	return output
}

type Repository struct {
	Type string `json:"type"`
	Url  string `json:"url"`
	// if a string rather than an object was given we'll store it in SingleLine
	SingleLine string
}

func (r Repository) ToString() string {
	if r.SingleLine != "" {
		return r.SingleLine
	}
	return r.Url
}

type Package struct {
	Config PackageConfig
	Path   string
	// for when something is linked to the global node_modules folder
	LinkedGlobally bool
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
