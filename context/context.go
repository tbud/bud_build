// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/tbud/bud/asset"
	"github.com/tbud/x/config"
	"github.com/tbud/x/log"
)

const (
	CONTEXT_CONFIG_TASK_KEY = "tasks"
	BUD_TASK_GROUP          = Group("bud")
)

var (
	ContextConfig config.Config
	Log           *log.Logger
)

func init() {
	var (
		homeDir string
		err     error
	)
	Log, err = log.New(nil)

	homeDir, err = HomeDir()
	ExitIfError(err)

	// get bud.conf from asset
	var reader io.ReadCloser
	reader, err = asset.Open("bud.conf")
	ExitIfError(err)
	ContextConfig, err = config.Read(reader)
	ExitIfError(err)
	reader.Close()

	// init config
	budConf := path.Join(homeDir, ".bud")
	if _, ferr := os.Stat(budConf); !os.IsNotExist(ferr) {
		var contextConfig config.Config
		contextConfig, err = config.Load(budConf)
		ExitIfError(err)
		ContextConfig.Merge("", contextConfig)
	}

	// init log
	Log, err = log.New(ContextConfig.SubConfig("log"))
	ExitIfError(err)

	// init asset log
	asset.InitLog(Log)
}

func TaskConfig(key string, value interface{}) error {
	if len(key) > 0 {
		key = fmt.Sprintf("%s.%s", CONTEXT_CONFIG_TASK_KEY, key)
	} else {
		return fmt.Errorf("key is empty")
	}
	return ContextConfig.Merge(key, value)
}

func ExitIfError(err error) {
	if err != nil {
		Log.Error(err.Error())
		os.Exit(1)
	}
}
