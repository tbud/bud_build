package seed

import (
	"fmt"
	. "github.com/tbud/bud/context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const Seed_Template_Suffix = ".seedtemplate"

type Step struct {
	Message string
}

type Seed interface {
	Name() string
	Description() string
	Start(args ...string) (*Step, error)
	NextStep(input string) (*Step, error)
}

var _seeds = []Seed{}
var _seedNames = map[string]int{}

func Register(seed Seed) {
	if seed == nil {
		panic("bud: Register seed is nil")
	}

	if _, exist := _seedNames[seed.Name()]; exist {
		panic(fmt.Errorf("Already has a seed named '%d'", seed.Name()))
	} else {
		_seedNames[seed.Name()] = len(_seeds)
		_seeds = append(_seeds, seed)
	}
}

func First() Seed {
	if len(_seeds) == 0 {
		return nil
	} else {
		return _seeds[0]
	}
}

func Each(fun func(Seed) error) error {
	for _, seed := range _seeds {
		err := fun(seed)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetSeed(name string) Seed {
	if index, exist := _seedNames[name]; exist {
		return _seeds[index]
	} else {
		return nil
	}
}

func CreateArchetype(destDir, srcDir string, data interface{}) error {
	var archetypeDir string
	// check seed dir wether or not a link
	fi, err := os.Lstat(srcDir)
	if err == nil && fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		if archetypeDir, err = os.Readlink(srcDir); err != nil {
			Log.Error("%v", err)
			return fmt.Errorf("Read link err %s", srcDir)
		}
	} else {
		archetypeDir = srcDir
	}

	// check seed archetype dir is exist
	if _, err := os.Stat(archetypeDir); err != nil {
		if os.IsNotExist(err) {
			Log.Error("%v", err)
			return fmt.Errorf("Seed archetype not exist: %s", archetypeDir)
		}
	}

	if err = os.MkdirAll(destDir, 0777); err != nil {
		Log.Error("%v", err)
		return fmt.Errorf("Failed to create directory: %s", destDir)
	}

	err = filepath.Walk(archetypeDir, func(path string, info os.FileInfo, err error) error {
		relSrcPath := strings.TrimLeft(path[len(archetypeDir):], string(os.PathSeparator))
		destPath := filepath.Join(destDir, relSrcPath)

		if strings.HasPrefix(relSrcPath, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			err = os.MkdirAll(destPath, 0777)
			if !os.IsNotExist(err) {
				return err
			}
			return nil
		}

		if strings.HasSuffix(relSrcPath, Seed_Template_Suffix) {
			return copyTemplateFile(destPath[:len(destPath)-len(Seed_Template_Suffix)], path, data)
		}

		return copyFile(destPath, path)
	})

	if err != nil {
		Log.Error("%v", err)
	}
	return err
}

func copyFile(destFile, srcFile string) (err error) {
	var dst, src *os.File
	if dst, err = os.Create(destFile); err != nil {
		return
	}

	if src, err = os.Open(srcFile); err != nil {
		return
	}

	if _, err = io.Copy(dst, src); err != nil {
		return
	}

	if err = src.Close(); err != nil {
		return
	}

	return dst.Close()
}

func copyTemplateFile(destFile, srcFile string, data interface{}) (err error) {
	var temp *template.Template
	if temp, err = template.ParseFiles(srcFile); err != nil {
		return err
	}

	var dst *os.File
	if dst, err = os.Create(destFile); err != nil {
		return err
	}

	if err = temp.Execute(dst, data); err != nil {
		return err
	}

	return dst.Close()
}
