package asset

import (
	"fmt"
	. "github.com/tbud/bud/context"
	"os"
	"path/filepath"
)

type AssetTask struct {
	Package  string
	BaseDir  string
	Includes []string
	Excludes []string
	Output   string
	Compress bool

	outputFilepath string
}

func init() {
	var err error
	assetTask := AssetTask{
		Package:  "main",
		Includes: []string{"**"},
		Output:   "./assets.go",
		Compress: true,
	}

	assetTask.BaseDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	Task("tobin", &assetTask)
}

func (a *AssetTask) Execute() error {
	err := a.Validate()
	if err != nil {
		return err
	}

	fmt.Println(a)
	return nil
}

func (a *AssetTask) Validate() error {
	if len(a.Package) == 0 {
		return fmt.Errorf("Missing package name")
	}

	var err error
	if !filepath.IsAbs(a.BaseDir) {
		a.BaseDir, err = filepath.Abs(a.BaseDir)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(a.BaseDir); os.IsNotExist(err) {
		return fmt.Errorf("Base dir : '%s' not exist", a.BaseDir)
	}

	if len(a.Output) == 0 {
		a.Output = "assets.go"
	}

	a.outputFilepath = filepath.Join(a.BaseDir, a.Output)
	stat, err := os.Lstat(a.outputFilepath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("Outpput path: %v", err)
		}

		dir, _ := filepath.Split(a.outputFilepath)
		if dir != "" {
			err = os.MkdirAll(dir, 0744)

			if err != nil {
				return fmt.Errorf("Create output directory: %v", err)
			}
		}
	}

	if stat != nil && stat.IsDir() {
		return fmt.Errorf("Output path is a directory.")
	}

	for _, include := range a.Includes {
		file := filepath.Join(a.BaseDir, include)
		matches, err := filepath.Glob(file)
		if err != nil {
			return err
		}
		fmt.Println("******************")
		fmt.Println(matches)
	}

	return nil
}
