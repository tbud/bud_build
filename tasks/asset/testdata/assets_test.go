// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testdata

import (
	"github.com/tbud/bud/tasks/asset"
	"io/ioutil"
	"strings"
	"testing"
)

func TestAssetRead(t *testing.T) {
	a, err := asset.Open("asset.go")
	if err != nil {
		t.Error(err)
	}

	var buf []byte
	buf, err = ioutil.ReadAll(a)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(string(buf), "asset.Register") {
		t.Errorf("asset format error")
	}
}

func TestAssetStat(t *testing.T) {
	a, err := asset.Stat("asset.go")
	if err != nil {
		t.Error(err)
	}

	if a.Name() != "asset.go" {
		t.Errorf("want asset.go get %s", a.Name())
	}
}
