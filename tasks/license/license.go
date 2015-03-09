package license

import (
	"bytes"
	"errors"
	. "github.com/tbud/bud/context"
	"github.com/tbud/x/path/selector"
	"io/ioutil"
	"os"
)

type licenseTask struct {
	BaseDir     string
	Patterns    []string
	LicenseHead string

	files []string
}

func init() {
	license := &licenseTask{
		Patterns: []string{"**/*.go"},
		LicenseHead: `// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

`}

	Task("license", BUD_TASK_GROUP, license, Usage("Add license head to files."))
}

func (l *licenseTask) Execute() (err error) {
	licenseBuf := []byte(l.LicenseHead)
	for _, filename := range l.files {
		if filebuf, err := ioutil.ReadFile(filename); err != nil {
			return err
		} else {
			if !bytes.HasPrefix(filebuf, licenseBuf) {
				if f, err := os.Open(filename); err != nil {
					return err
				} else {
					_, err = f.WriteAt(licenseBuf, 0)
					f.Close()
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (l *licenseTask) Validate() (err error) {
	if len(l.BaseDir) == 0 {
		if l.BaseDir, err = os.Getwd(); err != nil {
			return err
		}
	}

	if len(l.Patterns) == 0 {
		l.Patterns = []string{"**/*.go"}
	}

	if len(l.LicenseHead) == 0 {
		return errors.New("License head is empty.")
	}

	var s *selector.Selector
	if s, err = selector.New(l.Patterns...); err != nil {
		return err
	}

	if l.files, err = s.Matches(l.BaseDir); err != nil {
		return err
	}
	return nil
}