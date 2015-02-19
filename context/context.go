package context

import (
	"github.com/tbud/x/config"
	"github.com/tbud/x/log"
	"os"
)

var (
	Config config.Config
	Log    *log.Logger
)

func init() {
	var err error
	Log, err = log.New(nil)
	ExitIfError(err)
}

func ExitIfError(err error) {
	if err != nil {
		Log.Error(err.Error())
		os.Exit(1)
	}
}
