package cmd

import (
	"fmt"
	. "github.com/tbud/bud/common"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	BUD_DEFAULT_SEED_PATH = "github.com/tbud/sea"
)

func panicOnError(err error, msg string, args ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Abort: %s\n", err)
		fmt.Fprintf(os.Stderr, msg, args...)
		panic(err)
	}
}

func run(cmdargs ...interface{}) {
	cmdline := stringList(cmdargs...)
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		Log.Error("%v", err)
	}
}

func runOut(dir string, cmdargs ...interface{}) []byte {
	cmdline := stringList(cmdargs...)
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.Stderr.Write(out)
		Log.Error("%v", err)
		out = nil
	}
	return out
}

// matchPattern(pattern)(name) reports whether
// name matches pattern.  Pattern is a limited glob
// pattern in which '...' means 'any string' and there
// is no other special syntax.
func matchPattern(pattern string) func(name string) bool {
	re := regexp.QuoteMeta(pattern)
	re = strings.Replace(re, `\.\.\.`, `.*`, -1)
	// Special case: foo/... matches foo too.
	if strings.HasSuffix(re, `/.*`) {
		re = re[:len(re)-len(`/.*`)] + `(/.*)?`
	}
	reg := regexp.MustCompile(`^` + re + `$`)
	return func(name string) bool {
		return reg.MatchString(name)
	}
}

// stringList's arguments should be a sequence of string or []string values.
// stringList flattens them into a single []string.
func stringList(args ...interface{}) []string {
	var x []string
	for _, arg := range args {
		switch arg := arg.(type) {
		case []string:
			x = append(x, arg...)
		case string:
			x = append(x, arg)
		default:
			panic("stringList: invalid argument")
		}
	}
	return x
}
