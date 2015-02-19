package asset

import (
	"fmt"
	. "github.com/tbud/bud/context"
	"os"
	"path/filepath"
)

var _assets = map[string]Asset{}

func Register(assets []Asset) {

	if assets == nil {
		panic("bud: Register assets is nil")
	}

	for _, asset := range assets {
		if _, dup := _assets[asset.Name]; dup {
			panic("bud: Register called twice for asset " + asset.Name)
		}
	}
}

type AssetTask struct {
	Package  string
	BaseDir  string
	Includes []string
	Excludes []string
	Output   string
	Compress bool
}

func init() {
	var err error
	assetTask := AssetTask{
		Package:  "main",
		Includes: []string{"resource/**"},
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

	return nil
}

func (a *AssetTask) Validate() error {
	if len(a.Package) == 0 {
		return fmt.Errorf("Missing package name")
	}

	if !filepath.IsAbs(a.BaseDir) {
		a.BaseDir = filepath.Abs(a.BaseDir)
	}
	if _, err := os.Stat(a.BaseDir); os.IsNotExist(err) {
		return fmt.Errorf("Base dir : '%s' not exist", a.BaseDir)
	}

	if len(a.Output) == 0 {

	}

	return nil
}
