package commands

import "encoding/json"

// golang doesn't support union types, but fields like 'author' and 'repository' can actually be strings or objects so we'll need to keep that in mind when parsing

type PackageInput struct {
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
	BundleDependencies   bool              `json:"bundleDependencies"`
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
	Linked bool
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
	BundleDependencies   bool
}
