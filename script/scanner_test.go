// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package script

import (
	"path/filepath"
	"testing"
)

func TestScriptScanner(t *testing.T) {
	scan := scriptScanner{}

	file, err := filepath.Abs("scanner_test.bud")
	if err != nil {
		t.Error(err)
	}

	err = scan.checkValid(file)
	if err != nil {
		t.Error(err)
	}

	if len(scan.Imports) != 2 {
		t.Errorf("import num %d, got :%v", len(scan.Imports), scan.Imports)
	}

	if len(scan.Consts) != 2 {
		t.Errorf("const num %d, got :%v", len(scan.Consts), scan.Consts)
	}

	if len(scan.Funcs) != 3 {
		t.Errorf("func num %d, got :%v", len(scan.Funcs), scan.Funcs)
	}

	if len(scan.Types) != 2 {
		t.Errorf("type num %d, got :%v", len(scan.Types), scan.Types)
	}

	if len(scan.Vars) != 2 {
		t.Errorf("var num %d, got :%v", len(scan.Vars), scan.Vars)
	}

	if len(scan.Lines) != 5 {
		for _, line := range scan.Lines {
			println(line)
		}
		t.Errorf("line num %d, got :%v", len(scan.Lines), scan.Lines)
	}
}
