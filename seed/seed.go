package seed

import (
	"bufio"
	"fmt"
	. "github.com/tbud/bud/context"
	"os"
)

const Seed_Template_Suffix = ".seedtemplate"

type Step struct {
	Message string
}

type Seed interface {
	Name() string
	Description() string
	Start(args ...string) (*Step, error)
	NextStep(input string) (*Step, error)
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

func CreateSeed(args ...string) {
	if len(_seeds) == 0 {
		Log.Error("There is no seed to use.")
		return
	}

	initExitSign()

	seed := selectSeed()

	var step *Step
	var err error
	step, err = seed.Start(args...)
	if err != nil {
		Log.Error("%v", err)
		return
	}

	fmt.Println(step.Message)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		step, err = seed.NextStep(scanner.Text())
		if err != nil {
			Log.Error("%v", err)
			return
		}
		if step == nil {
			break
		} else {
			fmt.Println(step.Message)
		}
	}

	return
}

func initExitSign() {
}

func selectSeed() (seed Seed) {
	fmt.Println("Please select a seed to create.")
	each(func(seed Seed) error {
		fmt.Printf("\t%s\t-%s\n", seed.Name(), seed.Description())
		return nil
	})
	fmt.Printf("What is the seed name? [%s]\n", _seeds[0].Name())

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		name := scanner.Text()
		if len(name) == 0 {
			seed = _seeds[0]
		} else {
			seed = getSeed(name)
		}

		if seed == nil {
			fmt.Println("Please input a valid seed name.")
		} else {
			break
		}
	}

	return
}

func getSeed(name string) Seed {
	if index, exist := _seedNames[name]; exist {
		return _seeds[index]
	} else {
		return nil
	}
}

func each(fun func(Seed) error) error {
	for _, seed := range _seeds {
		err := fun(seed)
		if err != nil {
			return err
		}
	}
	return nil
}

const scriptTemplate = `
package main

import (
	"github.com/tbud/bud/seed"
{{ range $seed := .Seeds}}
	_ "{{ $seed }}"
{{ end }}
)

func main() {
	seed.CreateSeed()
}
`
