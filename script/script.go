package script

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/tbud/bud/builtin"
	"go/format"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

const scriptTemplate = `
package main

import (
	. "fmt"
	. "os"
	. "strings"
	. "strconv"
	. "github.com/tbud/x/config"
	. "github.com/tbud/bud/context"
	. "github.com/tbud/bud/builtin"
)

{{ range $import := .Imports}}
{{ $import }}
{{ end }}

{{ range $const := .Consts}}
{{ $const }}
{{ end }}

{{ range $var := .Vars}}
{{ $var }}
{{ end }}

{{ range $type := .Types}}
{{ $type }}
{{ end }}

{{ range $func := .Funcs}}
{{ $func }}
{{ end }}

func init() {
	Args = Args[1:]
	_ = Printf
	_ = Exit
	_ = Contains
	_ = Atoi
	_ = Config{}
	_ = Task
	_ = Exec
}

func main() {
	{{ range $line := .Lines}}
	{{ $line }}
	{{ end }}

	UseTasks()
	if len(Args) > 0 {
		for _, cmd := range Args {
			err := RunTask(cmd)
			if err != nil {
				Log.Error("%v", err)
				Exit(1)
			}
		}
	} else {
		err := RunTask("default")
		if err != nil {
			Log.Error("%v", err)
			Exit(1)
		}
	}
}
`

var (
	scriptDebug *bool
)

func init() {
	rand.Seed(time.Now().Unix())
}

func genDirAndFile() (tmpDir string, file string) {
	base, err := os.Getwd()
	if err != nil {
		base = os.TempDir()
	}

	for {
		tmpDir = filepath.Join(base, fmt.Sprintf(".budtmp.%08x", rand.Int63()))
		if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
			os.MkdirAll(tmpDir, 0700)
			return tmpDir, filepath.Join(tmpDir, "script.go")
		}
	}
}

func parseArgs(args ...string) (arg []string, err error) {
	flagSet := flag.FlagSet{}
	scriptDebug = flagSet.Bool("d", false, "show debug info")

	err = flagSet.Parse(args)
	if err != nil {
		return nil, err
	}

	return flagSet.Args(), nil
}

func genScriptBufFromTemplate(script string, debug bool, data interface{}) (buf []byte, err error) {
	templ, err := template.New("").Parse(script)
	if err != nil {
		return
	}

	var scriptBuf = bytes.Buffer{}
	err = templ.Execute(&scriptBuf, data)
	if err != nil {
		return
	}

	if debug {
		buf, err = format.Source(scriptBuf.Bytes())
		if err != nil {
			fmt.Println(scriptBuf.String())
			return
		}
	} else {
		buf = scriptBuf.Bytes()
	}

	return
}

func RunScript(script string, debug bool, data interface{}, args ...string) error {
	scriptBuf, err := genScriptBufFromTemplate(script, debug, data)
	if err != nil {
		return err
	}

	tempDir, scriptFile := genDirAndFile()
	err = ioutil.WriteFile(scriptFile, scriptBuf, 0600)
	if err != nil {
		return err
	}

	// timeB := time.Now()

	scriptExe := scriptFile + ".exe" // to be compatible with windows
	// the -ldflags -w could reduce exe size and faster build time
	err = builtin.Exec("go", "build", "-ldflags", "-w", "-o", scriptExe, scriptFile)
	if err != nil {
		return err
	}

	// println(time.Now().UnixNano() - timeB.UnixNano())

	err = builtin.Exec(scriptExe, args...)
	if err != nil {
		return err
	}

	if !debug {
		err = os.RemoveAll(tempDir)
		if err != nil {
			return err
		}
	}

	return nil

}

func Run(fileName string, args ...string) error {
	parsedArgs, err := parseArgs(args...)
	if err != nil {
		return err
	}

	scan := scriptScanner{}
	err = scan.checkValid(fileName)
	if err != nil {
		return err
	}

	return RunScript(scriptTemplate, *scriptDebug, scan, parsedArgs...)
}
