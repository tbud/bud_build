// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package seed

import (
	"fmt"
	. "github.com/tbud/bud/context"
	"github.com/tbud/x/io/ioutil"
	"os"
	"strings"
	"text/template"
)

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

	err = ioutil.Copy(destDir, archetypeDir, 0, func(dest, src string, srcInfo os.FileInfo) (skiped bool, err error) {
		if !srcInfo.IsDir() && strings.HasSuffix(src, Seed_Template_Suffix) {
			return true, copyTemplateFile(dest[:len(dest)-len(Seed_Template_Suffix)], src, data)
		}
		return false, nil
	})

	if err != nil {
		Log.Error("%v", err)
	}
	return err
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
