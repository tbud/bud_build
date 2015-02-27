package script

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
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
	//Args = Args[1:]
	_ = Printf
	_ = Exit
	_ = Contains
	_ = Abs
	_ = Atoi
	_ = Config{}
}

func main() {
{{ range $line := .Lines}}
{{ $line }}
{{ end }}
}
`

func init() {
	rand.Seed(time.Now().Unix())
}

func genDirAndFile(fileName string) (tmpDir string, file string) {
	base, file := filepath.Split(fileName)

	if !strings.HasSuffix(file, ".go") {
		file = file + ".go"
	}

	for {
		tmpDir = filepath.Join(base, fmt.Sprintf("%08x", rand.Int63()))
		if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
			os.MkdirAll(tmpDir, 0700)
			return tmpDir, filepath.Join(tmpDir, file)
		}
	}
}

func Run(fileName string) error {
	scan := scriptScanner{}
	err := scan.checkValid(fileName)
	if err != nil {
		return err
	}

	templ, err := template.New("").Parse(scriptTemplate)
	if err != nil {
		return err
	}

	scriptBuf := bytes.Buffer{}
	err = templ.Execute(&scriptBuf, scan)
	if err != nil {
		return err
	}

	buf, err := format.Source(scriptBuf.Bytes())
	if err != nil {
		fmt.Println(scriptBuf.String())
		return err
	}

	tempDir, scriptFile := genDirAndFile(fileName)
	err = ioutil.WriteFile(scriptFile, buf, 0600)
	if err != nil {
		return err
	}

	gocmd, err := exec.LookPath("go")
	if err != nil {
		return err
	}

	scriptExe := scriptFile + ".exe" // to be compatible with windows
	cmd := exec.Command(gocmd, "build", "-o", scriptExe, scriptFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err = cmd.Run(); err != nil {
		return err
	}

	exeCmd := exec.Command(scriptExe)
	exeCmd.Stdout = os.Stdout
	exeCmd.Stderr = os.Stderr
	exeCmd.Stdin = os.Stdin
	if err = exeCmd.Run(); err != nil {
		return err
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		return err
	}

	return nil
}
