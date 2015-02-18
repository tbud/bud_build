package plugins

import (
	"github.com/tbud/x/config"
)

type Plugin interface {
	Execute() error
	Validate() error
}

func AddBudPlugin(plugImport string, config config.Config) error {
	return nil
}
