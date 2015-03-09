// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package script

import (
	"path/filepath"
	"testing"
)

func TestScript(t *testing.T) {
	file, err := filepath.Abs("script_test.bud")
	if err != nil {
		t.Error(err)
	}

	err = Run(file)
	if err != nil {
		t.Error(err)
	}
}
