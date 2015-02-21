package context

import (
	"fmt"
	"github.com/tbud/x/config"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"unicode"
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
		err := configTask(taskName)
		if err != nil {
			return err
		}

		err = checkTaskValidate(taskName)
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
			conf := contextConfig.SubConfig(CONTEXT_CONFIG_TASK_KEY).SubConfig(t.packageName)
			if conf != nil {
				err := configExecutor(t.executor, conf)
				if err != nil {
					return err
				}
			}

			if t.config != nil {
				return configExecutor(t.executor, t.config)
			}
		}
		return nil
	})
}

func firstRuneToUpper(key string) string {
	rkey := []rune(key)
	rkey[0] = unicode.ToUpper(rkey[0])
	return string(rkey)
}

func configExecutor(executor Executor, conf config.Config) error {
	ev := reflect.ValueOf(executor)
	if ev.Kind() == reflect.Ptr {
		ev = ev.Elem()
	}

	return conf.EachKey(func(key string) error {
		value := ev.FieldByName(firstRuneToUpper(key))
		if value.IsValid() {
			switch value.Kind() {
			case reflect.Slice:
				if value.Type() == reflect.TypeOf([]string{}) {
					if ssv, ok := conf.Strings(key); ok {
						value.Set(reflect.ValueOf(ssv))
					}
				}
			case reflect.Bool:
				if bv, ok := conf.Bool(key); ok {
					value.SetBool(bv)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if iv, ok := conf.Int(key); ok {
					value.SetInt(int64(iv))
				}
			case reflect.String:
				if sv, ok := conf.String(key); ok {
					value.SetString(sv)
				}
			case reflect.Float32, reflect.Float64:
				if fv, ok := conf.Float(key); ok {
					value.SetFloat(fv)
				}
			}
		}
		return nil
	})
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
