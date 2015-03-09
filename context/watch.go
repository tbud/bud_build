package context

import (
	"fmt"
	"github.com/tbud/x/path/selector"
	"gopkg.in/fsnotify.v1"
	"os"
	"os/signal"
	"runtime"
	"sync"
)

var _watchs []*watch
var stopWatch chan bool

type watch struct {
	// watchFiles WatchFiles
	baseDir   string
	pselector *selector.Selector
	tasks     []string
	fun       func(event Event) error
	watcher   *fsnotify.Watcher
	lock      sync.Mutex
	exit      chan bool
}

type Event struct {
	fsnotify.Event
}

const (
	Op_Create fsnotify.Op = 1 << iota
	Op_Write
	Op_Remove
	Op_Rename
	Op_Chmod
)

type PatternsType []string
type BaseDir string

func Patterns(patterns ...string) PatternsType {
	return PatternsType(patterns)
}

func Watch(patterns PatternsType, args ...interface{}) {
	if len(patterns) > 0 {
		if len(args) == 0 {
			panic("watch must have related tasks or related function")
		}

		var err error
		w := &watch{baseDir: "."}

		w.pselector, err = selector.New([]string(patterns)...)
		if err != nil {
			panic(fmt.Errorf("parse pattern err: %v", err))
		}

		for i, arg := range args {
			switch value := arg.(type) {
			default:
				panic(fmt.Errorf("unknown args at arg[%d].", i+1))
			case func(event Event) error:
				w.fun = value
			case TasksType:
				w.tasks = []string(value)
			}
		}

		if err = w.initWatcher(); err != nil {
			panic(err)
		}

		_watchs = append(_watchs, w)
	}
}

func StartWatchs() (err error) {
	if len(_watchs) > 0 {

		for _, w := range _watchs {
			if err = w.beginWatch(); err != nil {
				Log.Error("%v", err)
				return err
			}
		}

		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt, os.Kill)
		<-ch

		StopWatchs()
	}

	return nil
}

func StopWatchs() error {
	if len(_watchs) > 0 {
		for _, w := range _watchs {
			w.exit <- true
		}
	}
	return nil
}

func (w *watch) initWatcher() (err error) {
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		Log.Error("%v", err)
		return err
	}

	w.watcher.Events = make(chan fsnotify.Event, 100)
	w.watcher.Errors = make(chan error, 10)

	return nil
}

func (w *watch) beginWatch() (err error) {
	// if err = w.addAllDir(); err != nil {
	// 	return err
	// }

	var matches []string
	if matches, err = w.pselector.Matches(w.baseDir); err != nil {
		return err
	}

	for _, match := range matches {
		w.watcher.Add(match)
	}

	go func() {
		for {
			select {
			case event := <-w.watcher.Events:
				w.event(Event{event})
			case err := <-w.watcher.Errors:
				Log.Warn("watcher catch an error: %v", err)
			case <-w.exit:
				w.watcher.Close()
				w.watcher = nil
				break
			}
		}
	}()

	return nil
}

func (w *watch) event(event Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	w.lock.Lock()
	defer w.lock.Unlock()

	if len(w.tasks) > 0 {
		for _, task := range w.tasks {
			if err = RunTask(task); err != nil {
				return err
			}
		}
	}

	if w.fun != nil {
		if err = w.fun(event); err != nil {
			return err
		}
	}

	return nil
}

func (w *watch) addAllDir() (err error) {
	var s *selector.Selector
	s, err = selector.New("d`**")
	if err != nil {
		return err
	}
	// add all dir
	var matches []string
	matches, err = s.Matches(w.baseDir)
	if err != nil {
		return err
	}

	for _, match := range matches {
		w.watcher.Add(match)
	}
	return nil
}
