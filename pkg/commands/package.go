package commands

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type PackageConfig struct {
	Name                 string
	Version              string
	License              string
	Description          string
	Homepage             string
	Main                 string
	Deprecated           bool
	Private              bool
	Files                []string
	Keywords             []string
	Os                   []string
	Cpu                  []string
	BundledDependencies  []string
	Scripts              map[string]string
	Directories          map[string]string
	Dependencies         map[string]string
	DevDependencies      map[string]string
	PeerDependencies     map[string]string
	OptionalDependencies map[string]string
	SortedDependencies   []*Dependency
	Engines              struct {
		Node string
		Npm  string
	}
	Repository   Repository
	Author       Author
	Contributors []Author
	Bugs         struct {
		Url string `json:"url"`
	}
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

func (p *Package) SortedDependencies() []*Dependency {
	deps := make([]*Dependency, 0, len(p.Config.Dependencies)+len(p.Config.DevDependencies)+len(p.Config.PeerDependencies)+len(p.Config.OptionalDependencies))

	type blah struct {
		kind   string
		depMap map[string]string
	}

	them := []blah{
		{
			kind:   "prod",
			depMap: p.Config.Dependencies,
		},
		{
			kind:   "dev",
			depMap: p.Config.DevDependencies,
		},
		{
			kind:   "peer",
			depMap: p.Config.PeerDependencies,
		},
		{
			kind:   "optional",
			depMap: p.Config.OptionalDependencies,
		},
	}

	for _, mapping := range them {
		depsForKind := make([]*Dependency, 0, len(mapping.depMap))
		for name, constraint := range mapping.depMap {
			depsForKind = append(depsForKind, &Dependency{
				Name:    name,
				Version: constraint,
				Kind:    mapping.kind,
			})
		}
		sort.Slice(depsForKind, func(i, j int) bool { return strings.Compare(depsForKind[i].Name, depsForKind[j].Name) < 0 })
		deps = append(deps, depsForKind...)
	}

	return deps
}

func (p *Package) SortedScripts() []*Script {
	scripts := make([]*Script, 0, len(p.Config.Scripts))
	for name, command := range p.Config.Scripts {
		scripts = append(scripts, &Script{Name: name, Command: command, ParentPackagePath: p.Path})
	}
	sort.Slice(scripts, func(i, j int) bool { return strings.Compare(scripts[i].Name, scripts[j].Name) < 0 })
	return scripts
}

func (p *Package) ConfigPath() string {
	return filepath.Join(p.Path, "package.json")
}

func (p *Package) ID() string {
	return fmt.Sprintf("package:%s", p.Path)
}

func (p *Package) Scoped() bool {
	return strings.HasPrefix(p.Config.Name, "@")
}
