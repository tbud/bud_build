package seed

import (
	"fmt"
)

type Step struct {
	Message string
}

type Seed interface {
	Name() string
	Description() string
	Start() *Step
	NextStep(input string) *Step
}

var _seeds = []Seed{}
var _seedNames = map[string]int{}

func Register(seed Seed) {
	if seed == nil {
		panic("bud: Register seed is nil")
	}

	if _, exist := _seedNames[seed.Name()]; exist {
		panic(fmt.Errorf("Already has a seed named '%d'", seed.Name()))
	} else {
		_seedNames[seed.Name()] = len(_seeds)
		_seeds = append(_seeds, seed)
	}
}

func First() Seed {
	if len(_seeds) == 0 {
		return nil
	} else {
		return _seeds[0]
	}
}

func Each(fun func(Seed) error) error {
	for _, seed := range _seeds {
		err := fun(seed)
		if err != nil {
			return err
		}
	}
	return nil
}

func FromName(seedName string) Seed {
	if index, exist := _seedNames[seedName]; exist {
		return _seeds[index]
	} else {
		return nil
	}
}
