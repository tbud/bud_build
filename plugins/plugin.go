package plugins

import (
	"github.com/tbud/x/config"
)

type Plugin interface {
	Execute() error
	Validate() error
}

var _plugins = []

func AddBudPlugin(plugImport string, config config.Config) error {
	
}
