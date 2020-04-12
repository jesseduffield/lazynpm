package commands

import "fmt"

type Script struct {
	Name              string
	Command           string
	ParentPackagePath string
}

func (s *Script) ID() string {
	return fmt.Sprintf("package:%s|script:%s", s.ParentPackagePath, s.Name)
}
