// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	. "github.com/tbud/x/builtin"
	. "github.com/tbud/x/config"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type task struct {
	name      string   // task name
	groupName string   // task group name
	tasks     []string // dependence tasks
	executor  Executor // if isn't nil, will execute after all dependence task called.
	config    Config   // task default config
	usageLine string   // usage info
}

var _tasks = map[string]*task{}
var _runningTask = []string{}
var _taskRunLock = sync.Mutex{}

type TasksType []string
type Usage string
type Group string

func Tasks(tasks ...string) TasksType {
	return TasksType(tasks)
}

func Task(name string, args ...interface{}) {
	if len(name) > 0 {
		if len(args) == 0 {
			panic("task must have dependence tasks or executor function")
		}

		t := task{name: name}
		for i, arg := range args {
			switch value := arg.(type) {
			default:
				panic(fmt.Errorf("unknown args at arg[%d].", i+1))
			case TasksType:
				t.tasks = []string(value)
			case Usage:
				t.usageLine = string(value)
			case Group:
				t.groupName = string(value)
			case func() error:
				ifPanic(t.executor != nil, fmt.Errorf("there is more than one executor in arg[%d].", i+1))
				t.executor = &defaultExecutor{runner: value}
			case Executor:
				ifPanic(t.executor != nil, fmt.Errorf("there is more than one executor in arg[%d].", i+1))
				t.executor = value
			case Config:
				ifPanic(t.config != nil, fmt.Errorf("there is more than one config in arg[%d].", i+1))
				t.config = value
			}
		}

		if len(t.groupName) == 0 {
			t.groupName = getTaskDefaultGroupName()
		}

		taskName := fmt.Sprintf("%s.%s", t.groupName, t.name)
		if _, exist := _tasks[taskName]; exist {
			panic("task name exist: " + taskName)
		}

		_tasks[taskName] = &t
		Log.Debug("Register task, name: %s", taskName)
	} else {
		panic("task name is empty")
	}
}

func RunTask(taskName string, config ...Config) (err error) {
	defer Catch(func(ierr interface{}) {
		switch value := ierr.(type) {
		case error:
			err = value
		}
		Log.Error("Catch error: %v", ierr)
	})

	if len(config) > 1 {
		return fmt.Errorf("Run task only receive one config param. There are %d config input.", len(config))
	}

	var conf Config = nil
	if len(config) == 1 {
		conf = config[0]
	}

	// lock task run
	_taskRunLock.Lock()
	defer _taskRunLock.Unlock()
	initTaskStack()

	if len(taskName) > 0 {
		err := configTask(taskName, conf)
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

func UseTasks(groupNames ...string) {
	if len(groupNames) == 0 {
		setTaskToDefault(getTaskDefaultGroupName())
	} else {
		for _, groupName := range groupNames {
			setTaskToDefault(groupName)
		}
	}
}

func initTaskStack() {
	if len(_runningTask) > 0 {
		Log.Error("task stack is not empty: %#v", _runningTask)

		_runningTask = _runningTask[:0]
	}
}

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

func getTaskDefaultGroupName() string {
	_, file, _, ok := runtime.Caller(2)
	if !ok {
		file = "???"
	}

	baseDir := filepath.Dir(file)
	baseDir = filepath.Base(baseDir)
	return baseDir
}

func configTask(taskName string, config Config) (err error) {
	return walkTask(taskName, func(t *task) error {
		if t.executor != nil {
			// task default config is the first config layer
			if t.config != nil {
				if err = t.config.SetStruct(t.executor); err != nil {
					return err
				}
			}

			// context group config is the second config layer
			conf := ContextConfig.SubConfig(CONTEXT_CONFIG_TASK_KEY).SubConfig(t.groupName)
			if conf != nil {
				if err = conf.SetStruct(t.executor); err != nil {
					return err
				} else {
					// context task config is the third config layer
					conf = conf.SubConfig(t.name)
					if conf != nil {
						if err = conf.SetStruct(t.executor); err != nil {
							return err
						}
					}
				}
			}

			// run task config is the last config layer
			if config != nil {
				return config.SetStruct(t.executor)
			}
			return nil
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
			Log.Debug("execute task '%s'.\nexecutor: %+v", t.name, t.executor)
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
		if len(task.tasks) > 0 {
			for _, depTask := range task.tasks {
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

func setTaskToDefault(groupName string) {
	packagePrefix := groupName + "."
	packagePrefixLen := len(packagePrefix)
	for taskName, task := range _tasks {
		if strings.HasPrefix(taskName, packagePrefix) {
			name := taskName[packagePrefixLen:]
			_tasks[name] = task
		}
	}
}
