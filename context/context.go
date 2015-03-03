package context

import (
	"fmt"
	"github.com/tbud/bud/asset"
	"github.com/tbud/x/config"
	"github.com/tbud/x/log"
	"os"
	"os/user"
	"path"
)

const (
	CONTEXT_CONFIG_TASK_KEY = "tasks"
	BUD_TASK_PACKAGE        = Group("bud")
)

var (
	contextConfig config.Config
	Log           *log.Logger
)

func init() {
	var err error

	currentUser, uerr := user.Current()
	ExitIfError(uerr)

	// get bud.conf from asset
	// init config
	budConf := path.Join(currentUser.HomeDir, ".bud")
	if _, ferr := os.Stat(budConf); !os.IsNotExist(ferr) {
		contextConfig, err = config.Load(budConf)
		ExitIfError(err)
	} else {
		contextConfig = config.Config{}
	}

	// init log
	Log, err = log.New(contextConfig.SubConfig("log"))
	ExitIfError(err)

	// init asset log
	asset.InitLog(Log)
}

func ContextConfig(conf config.Config) error {
	return contextConfig.Merge("", conf)
}

func TaskConfig(key string, value interface{}) error {
	if len(key) > 0 {
		key = fmt.Sprintf("%s.%s", CONTEXT_CONFIG_TASK_KEY, key)
	} else {
		key = CONTEXT_CONFIG_TASK_KEY
	}
	return contextConfig.Merge(key, value)
}

func ExitIfError(err error) {
	if err != nil {
		Log.Error(err.Error())
		os.Exit(1)
	}
}
