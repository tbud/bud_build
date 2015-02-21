package context

import (
	"github.com/tbud/x/config"
	"github.com/tbud/x/log"
	"os"
	"os/user"
	"path"
)

const (
	CONTEXT_CONFIG_TASK_KEY = "tasks"
)

var (
	contextConfig config.Config
	Log           *log.Logger
)

func init() {
	var err error
	Log, err = log.New(nil)
	ExitIfError(err)

	currentUser, uerr := user.Current()
	ExitIfError(uerr)

	budConf := path.Join(currentUser.HomeDir, ".bud")
	if _, ferr := os.Stat(budConf); !os.IsNotExist(ferr) {
		contextConfig, err = config.Load(budConf)
		ExitIfError(err)
	}
}

func ExitIfError(err error) {
	if err != nil {
		Log.Error(err.Error())
		os.Exit(1)
	}
}
