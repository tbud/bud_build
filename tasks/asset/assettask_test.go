// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asset

import (
	"testing"

	. "github.com/tbud/bud/context"
	. "github.com/tbud/x/config"
)

func TestAsset(t *testing.T) {
	TaskConfig("bud.asset", Config{
		"patterns": []string{"*.go"},
		"output":   "testdata/assets.go",
		"package":  "testdata",
	})
	// TaskConfig("asset.tobin.baseDir", "/Users/mind/gogo/src/github.com/tbud/x")

	// time.Sleep(30 * time.Second)
	UseTasks("bud")

	err := RunTask("asset")
	if err != nil {
		t.Error(err)
	}
}
