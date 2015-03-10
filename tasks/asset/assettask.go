// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asset

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	. "github.com/tbud/bud/asset"
	. "github.com/tbud/bud/context"
	"github.com/tbud/x/path/selector"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type AssetTask struct {
	Package  string
	BaseDir  string
	Patterns []string
	Output   string
	Compress bool

	outputFilepath string
	files          []string
}

func init() {
	var err error
	assetTask := AssetTask{
		Package:  "main",
		Patterns: []string{"**"},
		Output:   "./assets.go",
		Compress: true,
	}

	assetTask.BaseDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	Task("asset", BUD_TASK_GROUP, &assetTask, Usage("Package file into bin."))
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

		if a.outputFilepath == file {
			Log.Warn("files include output file: %s", file)
			continue
		}

		shortFile := strings.TrimPrefix(file, a.BaseDir)
		if len(shortFile) > 0 {
			shortFile = shortFile[1:]
			asset := Asset{
				N:  shortFile,
				D:  fi.IsDir(),
				S:  int(fi.Size()),
				M:  fi.Mode(),
				MT: fi.ModTime(),
			}
			assets = append(assets, asset)
		}
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

	if len(a.Patterns) == 0 {
		a.Patterns = []string{"**"}
	}

	var s *selector.Selector
	s, err = selector.New(a.Patterns...)
	if err != nil {
		return err
	}

	a.files, err = s.Matches(a.BaseDir)
	if err != nil {
		a.files = nil
		return err
	}

	return nil
}

const assetTemplate = `// GENERATED CODE - DO NOT EDIT
package {{ .assetTask.Package }}

import (
	"time"
	"github.com/tbud/bud/asset"
)	

func init() {
	asset.Register([]asset.Asset{
		{{ range $asset := .assets }}{
			N: "{{ $asset.N }}",
			D: {{ $asset.D }},
			{{if not $asset.D}}Z: []byte("{{ zipFile $asset.N $.assetTask.BaseDir | printf "%s" }}"),{{end}}
			S: {{ $asset.S }},
			M: {{ printf "%d" $asset.M }},
			MT : time.Unix({{ $asset.MT.Unix }}, 0),
		},{{ end }}
	})
}
`
