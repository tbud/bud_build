package context

import (
	"fmt"
	"github.com/tbud/x/config"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type task struct {
	name        string
	packageName string
	depends     []string
	executor    Executor
	config      config.Config
	usageLine   string
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

var _tasks = map[string]*task{}

var _runningTask = []string{}

func pushTask(taskName string) error {
	if len(_runningTask) == 0 {
		_runningTask = append(_runningTask, taskName)
	} else {
		for _, t := range _runningTask {
			if t == taskName {
				return fmt.Errorf("Already run task: %s, there is a recursion call. Call sequence:%v", t, _runningTask)
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

func ifPanic(bPanic bool, err error) {
	if bPanic {
		panic(err)
	}
}

func Task(name string, args ...interface{}) {
	if len(name) > 0 {
		if len(args) == 0 {
			panic("task must have dependence or executor function")
		}

		t := task{name: name}
		for i, arg := range args {
			switch value := arg.(type) {
			default:
				panic(fmt.Errorf("unknown args at arg[%d].", i+1))
			case Depends:
				t.depends = []string(value)
			case UsageLine:
				t.usageLine = string(value)
			case PackageName:
				t.packageName = string(value)
			case func() error:
				ifPanic(t.executor != nil, fmt.Errorf("there is more than one executor in arg[%d].", i+1))
				t.executor = &defaultExecutor{runner: value}
			case Executor:
				ifPanic(t.executor != nil, fmt.Errorf("there is more than one executor in arg[%d].", i+1))
				t.executor = value
			case config.Config:
				ifPanic(t.config != nil, fmt.Errorf("there is more than one config in arg[%d].", i+1))
				t.config = value
			}
		}

		if len(t.packageName) == 0 {
			t.packageName = getTaskDefaultPackageName()
		}

		taskName := fmt.Sprintf("%s.%s", t.packageName, t.name)
		if _, exist := _tasks[taskName]; exist {
			panic("task name exist: " + taskName)
		}

		_tasks[taskName] = &t
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

func RunTask(taskName string) error {
	if len(taskName) > 0 {
		err := checkTaskValidate(taskName)
		if err != nil {
			return err
		}

		return executeTask(taskName)
	} else {
		return fmt.Errorf("task name is empty")
	}
}

func configTask(taskName string) error {
	return walkTask(taskName, func(t *task) error {
		if t.executor != nil {
			conf := Config.SubConfig("tasks").SubConfig(t.packageName)
			return configExecutor(t.executor, conf, t.config)
		}
		return nil
	})
}

func configExecutor(executor Executor, contextConf config.Config, conf config.Config) error {
	ev := reflect.ValueOf(executor)
	if ev.Kind() == reflect.Ptr {
		ev = ev.Elem()
	}

	for i := 0; i < ev.NumField(); i++ {
		efv := ev.Field(i)
		name := 
		switch sf.Type.Kind() {

		}
	}

	return nil
}

func checkTaskValidate(taskName string) error {
	return walkTask(taskName, func(t *task) error {
		if t.executor != nil {
			return t.executor.Validate()
		}
		return nil
	})
}

func executeTask(taskName string) error {
	return walkTask(taskName, func(t *task) error {
		if t.executor != nil {
			return t.executor.Execute()
		}
		return nil
	})
}

func walkTask(taskName string, doTask func(t *task) error) error {
	err := pushTask(taskName)
	if err != nil {
		return err
	}
	defer popTask()

	if task, exist := _tasks[taskName]; exist {
		if len(task.depends) > 0 {
			for _, depTask := range task.depends {
				err := walkTask(depTask, doTask)
				if err != nil {
					return err
				}
			}
		}

		return doTask(task)
	} else {
		return fmt.Errorf("Could not find task: %s", taskName)
	}
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
