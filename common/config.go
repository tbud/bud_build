package common

import (
	"github.com/tbud/x/config"
	"github.com/tbud/x/log"
	"os"
)

var (
	Config *config.Config
	Log    *log.Logger
)

func init() {
	var err error
	if _, err = os.Stat(".bud"); err == nil {
		Config, err = config.Load(".bud")
		ExitIfError(err)
	}

	Log, err = log.New(Config.SubConfig("logger"))
	ExitIfError(err)
}
