package main

import (
	"fmt"
	"github.com/tbud/bud/script"
	"github.com/tbud/bud/seed"
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	// flag.Parse()
	// args := flag.Args()
	args := os.Args[1:]

	printIcon()

	// Diagnose common mistake: GOPATH==GOROOT.
	// This setting is equivalent to not setting GOPATH at all,
	// which is not what most people want when they do it.
	if gopath := os.Getenv("GOPATH"); gopath == runtime.GOROOT() {
		fmt.Fprintf(os.Stderr, "warning: GOPATH set to GOROOT (%s) has no effect\n", gopath)
	} else {
		for _, p := range filepath.SplitList(gopath) {
			// Note: using HasPrefix instead of Contains because a ~ can appear
			// in the middle of directory elements, such as /tmp/git-1.8.2~rc3
			// or C:\PROGRA~1. Only ~ as a path prefix has meaning to the shell.
			if strings.HasPrefix(p, "~") {
				fmt.Fprintf(os.Stderr, "bud: GOPATH entry cannot start with shell metacharacter '~': %q\n", p)
				os.Exit(2)
			}
			if build.IsLocalImport(p) {
				fmt.Fprintf(os.Stderr, "bud: GOPATH entry is relative; must be absolute path: %q.\nRun 'go help gopath' for usage.\n", p)
				os.Exit(2)
			}
		}
	}

	if len(args) == 0 {
		runBud(nil)
		return
	}

	switch args[0] {
	case "help":
		fmt.Print(helpUsage)
		println(helpFootprint)
		return
	case "new":
		seed.Run(args[1:]...)
		return
	}

	runBud(args)
}

var budIcon = []string{
	"",
	"\x1B[31m                )         (        \x1B[39m",
	"\x1B[31m             ( /(    (    )\\ )     \x1B[39m",
	"\x1B[31m             )\\())  ))\\  (()/(     \x1B[39m",
	"\x1B[35m            ((_)\\  /((_)  ((_))      \x1B[39m",
	"\x1B[32m            | |\x1B[39m\x1B[33m(_)(_))( \x1B[39m\x1B[32m  _| |       \x1B[39m",
	"\x1B[32m            | '_ \\| || |/ _  |       \x1B[39m",
	"\x1B[32m            |_.__/ \\_._|\\__._|     \x1B[39m",
	"",
	"Bud aim to be a full stack develop tool for Go language.",
}

var helpUsage = `Welcome to bud 0.0.5!

The commands are available:
---------------------------
    new         create a bud application from seed

`

var helpNoneBudDir = `Use "bud new" to create a new bud application in the current directory,
or go to an existing application and launch the development console using "bud".
`

var helpFootprint = "You can also browse informations at http://www.tbud.io.\n"

func printIcon() {
	for _, icon := range budIcon {
		println(icon)
	}
}

func runBud(args []string) {
	var budFile string
	var err error
	useDefaultBudFile := true
	if len(args) > 0 {
		budFile, err = filepath.Abs(args[0])
		if err == nil {
			if fi, err := os.Stat(budFile); !os.IsNotExist(err) && !fi.IsDir() {
				useDefaultBudFile = false
			}
		}
	}

	if useDefaultBudFile {
		budFile, err = filepath.Abs("build.bud")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Get build.bud file error: %v", err)
			return
		}

		if _, err = os.Stat(budFile); os.IsNotExist(err) {
			fmt.Fprint(os.Stderr, "\n\x1B[31mThis is not a bud application!\x1B[39m\n\n")
			println(helpNoneBudDir)
			println(helpFootprint)
			return
		}
	} else {
		args = args[1:]
	}

	script.Run(budFile, args...)
}
