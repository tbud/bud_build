// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asset

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/tbud/x/log"
	"io"
	"os"
	"time"
)

var _assets = map[string]Asset{}

var _logger *log.Logger

func init() {
	_logger, _ = log.New(nil)
}

// call this function in 'init' func
func InitLog(logger *log.Logger) {
	_logger = logger
}

func Register(assets []Asset) {
	if assets == nil {
		panic("bud: Register assets is nil")
	}

	for _, asset := range assets {
		if _, dup := _assets[asset.N]; dup {
			panic("bud: Register called twice for asset " + asset.N)
		} else {
			_assets[asset.N] = asset
		}
	}
}

func Open(name string) (rc io.ReadCloser, err error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("asset open file name empty")
	}

	if asset, exist := _assets[name]; exist {
		_logger.Info("Open '%s' from asset.", name)
		return &asset, nil
	} else {
		_logger.Info("Open '%s' from file system.", name)
		return os.Open(name)
	}
}

func Stat(name string) (fi os.FileInfo, err error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("asset open file name empty")
	}

	if asset, exist := _assets[name]; exist {
		_logger.Info("Load '%s' state from asset.", name)
		return &asset, nil
	} else {
		_logger.Info("Load '%s' state from file system.", name)
		return os.Stat(name)
	}
}

type Asset struct {
	N   string        // save name
	Z   []byte        // save compressed data
	S   int           // origin file size
	D   bool          // save is dir
	M   os.FileMode   // file mode
	MT  time.Time     // file mod time
	Buf *bytes.Buffer // used by read method
}

func (a *Asset) Read(p []byte) (n int, err error) {
	if a.IsDir() {
		return 0, nil
	}

	if a.Buf == nil {
		gz, err := gzip.NewReader(bytes.NewBuffer(a.Z))
		if err != nil {
			return 0, err
		}

		a.Buf = bytes.NewBuffer(nil)
		_, err = io.Copy(a.Buf, gz)
		gz.Close()
		if err != nil {
			return 0, err
		}

		if a.Buf.Len() != a.S {
			return 0, fmt.Errorf("Asset unzip file %s error.", a.N)
		}
	}

	return a.Buf.Read(p)
}

func (a *Asset) Close() error {
	return nil
}

func (a *Asset) Name() string {
	return a.N
}

func (a *Asset) Size() int64 {
	return int64(a.S)
}

func (a *Asset) Mode() os.FileMode {
	return a.M
}

func (a *Asset) ModTime() time.Time {
	return a.MT
}

func (a *Asset) IsDir() bool {
	return a.D
}
func (a *Asset) Sys() interface{} {
	return nil
}
