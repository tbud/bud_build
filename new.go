package main

import (
	// "bufio"
	// "bytes"
	// "fmt"
	// "os"
	// "os/exec"
	termbox "github.com/nsf/termbox-go"
)

var cmdNew = &Command{
	UsageLine: "new [path]",
	Short:     "create a bud application from seed",
	Long: `
New creates a few files to get a new bud application running quickly.

It puts all of the files in the given path, taking the final element in
the path to be the app name.

The -s flag is an optional argument, provided the ability to create from a special seed.
The default seed is react.
    `,
}

var seedName = cmdNew.Flag.String("s", "tea", "")

var (
	srcRoot    string
	appPath    string
	appName    string
	basePath   string
	importPath string
)

func init() {
	cmdNew.Run = newCommand
}

func newCommand(cmd *Command, args []string) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

}

func newBud()
