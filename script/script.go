package script

import (
// "text/template"
)

const scriptTemplate = `
package main

import (
	. "fmt"
	. "os"
	. "strings"
	. "math"
	. "strconv"
	. "github.com/tbud/x/Config"
)

func init() {
	//Args = Args[1:]
	_ = Printf
	_ = Exit
	_ = Contains
	_ = Abs
	_ = Atoi
}

func main() {
	Println("hello world")
}
`

func Run(file string) error {
	return nil
}
