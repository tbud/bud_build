// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seed

import (
	"bufio"
	"fmt"
	. "github.com/tbud/bud/context"
	"github.com/tbud/bud/script"
	"os"
)

const Seed_Template_Suffix = ".seedtemplate"

type Prompt struct {
	Message string
}

type Seed interface {
	Name() string
	Description() string
	Start(args ...string) (*Prompt, error)
	NextStep(input string) (*Prompt, error)
	HasNext() bool
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

	seed := selectSeed()

	var prompt *Prompt
	var err error
	prompt, err = seed.Start(args...)
	if err != nil {
		Log.Error("%v", err)
		return
	}

	fmt.Println(prompt.Message)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		prompt, err = seed.NextStep(scanner.Text())
		if err != nil {
			Log.Error("%v", err)
			return
		}
		if prompt == nil {
			break
		} else {
			fmt.Println(prompt.Message)
			if seed.HasNext() {
				fmt.Print("> ")
			} else {
				break
			}
		}
	}

	return
}

func selectSeed() (seed Seed) {
	printSeedList()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		name := scanner.Text()
		if len(name) == 0 {
			seed = _seeds[0]
		} else {
			seed = getSeed(name)
		}

		if seed == nil {
			Log.Error("Please input a valid seed name.")
			printSeedList()
		} else {
			break
		}
	}

	return
}

func printSeedList() {
	fmt.Println("Please select a seed to create.")
	each(func(seed Seed) error {
		fmt.Printf("\t%s\t- %s\n", seed.Name(), seed.Description())
		return nil
	})
	fmt.Printf("What is the seed name? [%s]\n> ", _seeds[0].Name())
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

func Run(args ...string) {
	err := script.RunScript(scriptTemplate, false, ContextConfig.StringsDefault("seed", nil), args...)
	if err != nil {
		Log.Error("%v", err)
	}
}

const scriptTemplate = `
package main

import (
	"github.com/tbud/bud/seed"
	"os"
{{ range $seed := .}}
	_ "{{ $seed }}"
{{ end }}
)

func main() {
	seed.CreateSeed(os.Args[1:]...)
}
`
