package plugins

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type task struct {
	name        string
	packageName string
	depends     []string
	usageLine   string
	runner      func() error
}

type Depends []string
type UsageLine string
type PackageName string

func Depend(tasks ...string) Depends {
	return Depends(tasks)
}

func Usage(line string) UsageLine {
	return UsageLine(line)
}

func Package(name string) PackageName {
	return PackageName(name)
}

var _tasks = map[string]task{}

var _runningTask = []string{}

func pushTask(taskName string) error {
	if len(_runningTask) == 0 {
		_runningTask = append(_runningTask, taskName)
	} else {
		for _, t := range _runningTask {
			if t == taskName {
				return errors.New(fmt.Sprintf("Already run task: %s, there is a recursion call. Call sequence:%v", t, _runningTask))
			}
		}
		_runningTask = append(_runningTask, taskName)
	}
	return nil
}

func popTask() {
	if len(_runningTask) > 0 {
		_runningTask = _runningTask[:len(_runningTask)-1]
	}
}

func Task(name string, args ...interface{}) {
	if len(name) > 0 {
		if len(args) == 0 {
			panic("task must have dependence or run function")
		}

		t := task{name: name}
		for i, arg := range args {
			switch value := arg.(type) {
			default:
				panic("unknown args on " + strconv.Itoa(i))
			case Depends:
				t.depends = []string(value)
			case []string:
				t.depends = value
			case UsageLine:
				t.usageLine = string(value)
			case PackageName:
				t.packageName = string(value)
			case func() error:
				t.runner = value
			}
		}

		if len(t.packageName) == 0 {
			t.packageName = getTaskDefaultPackageName()
		}

		taskName := fmt.Sprintf("%s.%s", t.packageName, t.name)
		if _, exist := _tasks[taskName]; exist {
			panic("task name exist: " + taskName)
		}

		_tasks[taskName] = t
	} else {
		panic("task name is empty")
	}
}

func getTaskDefaultPackageName() string {
	_, file, _, ok := runtime.Caller(2)
	if !ok {
		file = "???"
	}

	baseDir := filepath.Dir(file)
	baseDir = filepath.Base(baseDir)
	return baseDir
}

func RunTask(taskName string, args ...string) error {
	if len(taskName) > 0 {
		err := pushTask(taskName)
		if err != nil {
			return err
		}
		defer popTask()

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
				return task.runner()
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

func TaskPackageToDefault(packageNames ...string) {
	if len(packageNames) == 0 {
		setTaskToDefault(getTaskDefaultPackageName())
	} else {
		for _, packageName := range packageNames {
			setTaskToDefault(packageName)
		}
	}
}

func setTaskToDefault(packageName string) {
	packagePrefix := packageName + "."
	packagePrefixLen := len(packagePrefix)
	for taskName, task := range _tasks {
		if strings.HasPrefix(taskName, packagePrefix) {
			name := taskName[packagePrefixLen:]
			_tasks[name] = task
		}
	}
}
