package plugins

import (
	"errors"
	"github.com/tbud/bud/context"
)

type task struct {
	name    string
	depends []string
	runner  func(context context.Context, args ...string) error
}

type Depends []string

var _tasks = map[string]task{}

var _runningTask = []string{}

func Task(taskName string, args ...interface{}) {
	if len(taskName) > 0 {
		if _, exist := _tasks[taskName]; exist {
			panic("task name exist: " + taskName)
		}

		if len(args) == 0 {
			panic("task must have dependence or run function")
		}

		t := task{name: taskName}
		for _, arg := range args {
			switch value := arg.(type) {
			case Depends:
				t.depends = []string(value)
			case []string:
				t.depends = value
			case func(context context.Context, args ...string) error:
				t.runner = value
			}
		}

		_tasks[taskName] = t
	} else {
		panic("task name is empty")
	}
}

func Depend(tasks ...string) Depends {
	return Depends(tasks)
}

func RunTask(taskName string, args ...string) error {
	if len(taskName) > 0 {
		if task, exist := _tasks[taskName]; exist {
			if len(task.depends) > 0 {
				for _, depTask := range task.depends {
					err := RunTask(depTask, args...)
					if err != nil {
						return err
					}
				}
			}

			if task.runner != nil {
				return task.runner(context.Context{}, args...)
			}

			return nil
		} else {
			panic("no task : " + taskName)
		}
	} else {
		panic("task name is empty")
	}

	return errors.New("Run task not reach here!")
}
