package builtin

import (
	"path/filepath"
)

type Paths []string

func FindFiles(pattern string) Paths {
	matches, err := filepath.Glob(pattern)
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
