package context

import (
	"errors"
)

type Executor interface {
	Execute() error
	Validate() error
}

type defaultExecutor struct {
	runner func() error
}

func (d *defaultExecutor) Execute() error {
	if d != nil && d.runner != nil {
		return d.runner()
	}
	return nil
}

func (d *defaultExecutor) Validate() error {
	if d == nil || d.runner == nil {
		return errors.New("Current object or runner is nil")
	}
	return nil
}

func (d *defaultExecutor) String() string {
	return "Default Executor"
}
