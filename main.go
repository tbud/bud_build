package main

import (
	"bytes"
	"flag"
	"fmt"
	. "github.com/tbud/bud/cmd"
	"go/build"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	if args[0] == "help" {
		help(args[1:])
		return
	}

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
			Exit()
			return
		}
	}

	fmt.Fprintf(os.Stderr, "bud: unknown subcommand %q\nRun 'bud help' for usage.\n", args[0])
	SetExitStatus(2)
	Exit()
}

var usageTemplate = `Bud is a full stack develop tool for Go language.

Usage:

	bud command [arguments]

The commands are:
{{range .}}{{if .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "bud help [command]" for more information about a command.

Additional help topics:
{{range .}}{{if not .Runnable}}
    {{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "bud help [topic]" for more information about that topic.

`

var helpTemplate = `{{if .Runnable}}usage: bud {{.UsageLine}}

{{end}}{{.Long | trim}}
`

var documentationTemplate = `// Copyright 2011 The Bud Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT THIS FILE. GENERATED BY mkdoc.sh.
// Edit the documentation in other files and rerun mkdoc.sh to generate this one.

/*
{{range .}}{{if .Short}}{{.Short | capitalize}}

{{end}}{{if .Runnable}}Usage:

	bud {{.UsageLine}}

{{end}}{{.Long | trim}}


{{end}}*/
package main
`

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

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, Commands)
}

func usage() {
	// special case "go test -h"
	if len(os.Args) > 1 && os.Args[1] == "test" {
		help([]string{"testflag"})
		os.Exit(2)
	}
	printUsage(os.Stderr)
	os.Exit(2)
}

// help implements the 'help' command.
func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		// not exit 2: succeeded at 'bud help'.
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: bud help command\n\nToo many arguments given.\n")
		os.Exit(2) // failed at 'bud help'
	}

	arg := args[0]

	// 'bud help documentation' generates doc.go.
	if arg == "documentation" {
		buf := new(bytes.Buffer)
		printUsage(buf)
		usage := &Command{Long: buf.String()}
		tmpl(os.Stdout, documentationTemplate, append([]*Command{usage}, Commands...))
		return
	}

	for _, cmd := range Commands {
		if cmd.Name() == arg {
			tmpl(os.Stdout, helpTemplate, cmd)
			// not exit 2: succeeded at 'bud help cmd'.
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run 'bud help'.\n", arg)
	os.Exit(2) // failed at 'bud help cmd'
}
