package context

import (
	"errors"
)

type defaultPlugin struct {
	runner func() error
}

func (d *defaultPlugin) Execute() error {
	if d != nil && d.runner != nil {
		return d.runner()
	}
	return nil
}

func (d *defaultPlugin) Validate() error {
	if d == nil || d.runner == nil {
		return errors.New("Current object or runner is nil")
	}
	return nil
}
