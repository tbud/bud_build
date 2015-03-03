package context

import (
	"fmt"
	"github.com/tbud/bud/asset"
	"github.com/tbud/x/config"
	"github.com/tbud/x/log"
	"io"
	"os"
	"os/user"
	"path"
)

const (
	CONTEXT_CONFIG_TASK_KEY = "tasks"
	BUD_TASK_PACKAGE        = Group("bud")
)

var (
	_contextConfig config.Config
	Log            *log.Logger
)

func init() {
	var err error

	currentUser, uerr := user.Current()
	ExitIfError(uerr)

	// get bud.conf from asset
	var reader io.ReadCloser
	reader, err = asset.Open("bud.conf")
	ExitIfError(err)
	_contextConfig, err = config.Read(reader)
	ExitIfError(err)
	reader.Close()

	// init config
	budConf := path.Join(currentUser.HomeDir, ".bud")
	if _, ferr := os.Stat(budConf); !os.IsNotExist(ferr) {
		var contextConfig config.Config
		contextConfig, err = config.Load(budConf)
		ExitIfError(err)
		_contextConfig.Merge("", contextConfig)
	}

	// init log
	Log, err = log.New(_contextConfig.SubConfig("log"))
	ExitIfError(err)

	// init asset log
	asset.InitLog(Log)
}

func ContextConfig(conf config.Config) error {
	return _contextConfig.Merge("", conf)
}

func TaskConfig(key string, value interface{}) error {
	if len(key) > 0 {
		key = fmt.Sprintf("%s.%s", CONTEXT_CONFIG_TASK_KEY, key)
	} else {
		key = CONTEXT_CONFIG_TASK_KEY
	}
	return _contextConfig.Merge(key, value)
}

func ExitIfError(err error) {
	if err != nil {
		Log.Error(err.Error())
		os.Exit(1)
	}
}
