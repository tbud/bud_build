package builtin

import (
	"os"
	"os/exec"
	"path/filepath"
)

func Exec(name string, args ...string) (err error) {
	if !filepath.IsAbs(name) {
		name, err = exec.LookPath(name)
		if err != nil {
			return
		}
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
