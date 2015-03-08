package builtin

import (
	"github.com/tbud/x/path/selector"
)

type Paths []string

func FindFiles(pattern string) Paths {
	s, err := selector.New(pattern)
	if err != nil {
		panic(err)
	}

	var matches []string
	matches, err = s.Matches(".")
	if err != nil {
		panic(err)
	}

	return matches
}

func (p Paths) Each(fun func(string) error) error {
	if fun != nil {
		for _, path := range p {
			err := fun(path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
