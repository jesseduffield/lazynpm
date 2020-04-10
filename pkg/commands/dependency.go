package commands

type Dependency struct {
	Name         string
	Version      string
	LinkPath     string
	Present      bool
	LocalVersion string
}

func (d *Dependency) Linked() bool {
	return d.LinkPath != ""
}
