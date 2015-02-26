package asset

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	. "github.com/tbud/bud/context"
	"github.com/tbud/x/container/set"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type AssetTask struct {
	Package  string
	BaseDir  string
	Includes []string
	Excludes []string
	Output   string
	Compress bool

	outputFilepath string
	files          []string
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

func zipFile(fileName string, baseDir string) (ret []byte) {
	ret = []byte{}
	fileAbs := filepath.Join(baseDir, fileName)
	openFile, err := os.Open(fileAbs)
	if err != nil {
		panic(err)
	}
	defer openFile.Close()

	buf := bytes.NewBuffer(nil)

	gzw := gzip.NewWriter(&HexWriter{Writer: buf})
	_, err = io.Copy(gzw, openFile)
	gzw.Close()

	if err != nil {
		panic(err)
	}

	ret = buf.Bytes()
	return
}

func (a *AssetTask) Execute() error {
	out, err := os.Create(a.outputFilepath)
	if err != nil {
		return err
	}
	defer out.Close()

	bfd := bufio.NewWriter(out)
	defer bfd.Flush()

	var assets []Asset
	for _, file := range a.files {
		var fi os.FileInfo
		fi, err = os.Stat(file)
		if err != nil {
			return err
		}

		shortFile := strings.TrimPrefix(file, a.BaseDir)[1:]

		asset := Asset{
			N:  shortFile,
			S:  int(fi.Size()),
			M:  fi.Mode(),
			MT: fi.ModTime(),
		}
		assets = append(assets, asset)
	}

	funcMap := template.FuncMap{
		"zipFile": zipFile,
	}

	var templ *template.Template
	templ, err = template.New("").Funcs(funcMap).Parse(assetTemplate)
	if err != nil {
		return err
	}

	templ.Execute(bfd, map[string]interface{}{
		"assetTask": a,
		"assets":    assets,
	})

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

	fileSet := set.NewStringSet()
	if err := checkFilePath(a.BaseDir, a.Includes, func(matches []string) error {
		fileSet.Union(matches...)
		return nil
	}); err != nil {
		return err
	}

	if err := checkFilePath(a.BaseDir, a.Excludes, func(matches []string) error {
		fileSet.Subtract(matches...)
		return nil
	}); err != nil {
		return err
	}

	a.files = fileSet.ToSeq()
	return nil
}

func checkFilePath(baseDir string, filePaths []string, fun func(matches []string) error) error {
	for _, filePath := range filePaths {
		var file string
		if filepath.IsAbs(filePath) {
			file = filePath
		} else {
			file = filepath.Join(baseDir, filePath)
		}

		matches, err := filepath.Glob(file)
		if err != nil {
			return err
		}

		return fun(matches)
	}
	return nil
}
