package commands

import "fmt"

type Tarball struct {
	Name string
	Path string
}

func (t *Tarball) ID() string {
	return fmt.Sprintf("tarball:%s", t.Path)
}
