package main

import (
	"flag"
	"fmt"
	"github.com/tbud/bud/script"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

// Commands lists the available commands and help topics.
// The order here is the order in which they are printed by 'go help'.
var Commands = []*Command{
	cmdNew,
}

// A Command is an implementation of a go command
// like go build or go fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'bud help' output.
	Short string

	// Long is the long message shown in the 'bud help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet

	// CustomFlags indicates that the command will do its own
	// flag parsing.
	CustomFlags bool
}

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

// Runnable reports whether the command can be run; otherwise
// it is a documentation pseudo-command such as importpath.
func (c *Command) Runnable() bool {
	return c.Run != nil
}

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

	if len(args) < 1 {
		runBud(nil)
		return
	}

	if args[0] == "help" {
		printUsage(os.Stdout)
		println(helpFootprint)
		return
	}

	for _, cmd := range Commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			cmd.Run(cmd, args)
			return
		}
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

var usageTemplate = `Welcome to bud 0.0.3!

The commands are available:
---------------------------{{range .}}{{if .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

`

var helpNoneBudDir = `Use "bud new" to create a new bud application in the current directory,
or go to an existing application and launch the development console using "bud".
`

var helpFootprint = "You can also browse informations at http://www.tbud.io.\n"

// tmpl executes the given template text on data, writing the result to w.
func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace, "capitalize": capitalize})
	template.Must(t.Parse(text))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

func printIcon() {
	for _, icon := range budIcon {
		println(icon)
	}
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, Commands)
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
