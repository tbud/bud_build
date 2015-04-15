package dist

import (
	"fmt"
	"path/filepath"

	. "github.com/tbud/bud/builtin"
	. "github.com/tbud/bud/context"
	"github.com/tbud/bud/script"
)

type DistTask struct {
	DistDir       string
	BinName       string
	BudScriptFile string
	GOOSs         []string
	GOARCHs       []string
	Debug         bool
}

func init() {
	distTask := &DistTask{
		DistDir:       "./budist",
		BinName:       "budbin",
		BudScriptFile: "./build.bud",
		GOOSs:         []string{"darwin", "linux"},
		GOARCHs:       []string{"amd64"},
		Debug:         false,
	}

	Task("dist", BUD_TASK_GROUP, distTask, Usage("Distributing bud script to bin for all valid platform."))
}

func (d *DistTask) Execute() (err error) {
	distDir := d.DistDir
	if !filepath.IsAbs(distDir) {
		if distDir, err = filepath.Abs(distDir); err != nil {
			return err
		}
	}

	distSrcFile := filepath.Join(distDir, "source.go")

	budFile := d.BudScriptFile
	if !filepath.IsAbs(budFile) {
		if budFile, err = filepath.Abs(budFile); err != nil {
			return err
		}
	}

	if err = script.GenScript(budFile, distSrcFile, d.Debug); err != nil {
		return err
	}

	for _, goos := range d.GOOSs {
		for _, goarch := range d.GOARCHs {
			binName := filepath.Join(distDir, fmt.Sprintf("%s_%s", goos, goarch), d.BinName)
			if goos == "windows" {
				binName += ".exe"
			}
			if err = Cmd("go", "build", "-ldflags", "-w", "-o", binName, distSrcFile).WithEnv("GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=1").Run(); err != nil {
				// if err = Cmd("go", "build", "-o", binName, distSrcFile).WithEnv("GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0").Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DistTask) Validate() error {
	if len(d.DistDir) == 0 {
		return fmt.Errorf("dist dir is empty")
	}

	if len(d.BinName) == 0 {
		return fmt.Errorf("bin name is empty")
	}

	if len(d.BudScriptFile) == 0 {
		return fmt.Errorf("Bud script file name is empty")
	}

	if len(d.GOOSs) == 0 {
		return fmt.Errorf("GOOSs is empty")
	}

	if len(d.GOARCHs) == 0 {
		return fmt.Errorf("GOARCHs is empty")
	}

	return nil
}
