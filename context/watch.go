package context

import (
	"fmt"
	"github.com/tbud/x/path/selector"
	"gopkg.in/fsnotify.v1"
	"os"
	"os/signal"
	"runtime"
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
	exit      chan bool
	skipOp    fsnotify.Op
}

const (
	Op_Create fsnotify.Op = 1 << iota
	Op_Write
	Op_Remove
	Op_Rename
	Op_Chmod
)

type Event struct {
	fsnotify.Event
}

func (e *Event) IsCreate() bool {
	return e.Op&fsnotify.Create == fsnotify.Create
}

func (e *Event) IsWrite() bool {
	return e.Op&fsnotify.Write == fsnotify.Write
}

func (e *Event) IsRemove() bool {
	return e.Op&fsnotify.Remove == fsnotify.Remove
}

func (e *Event) IsRename() bool {
	return e.Op&fsnotify.Rename == fsnotify.Rename
}

func (e *Event) IsChmod() bool {
	return e.Op&fsnotify.Chmod == fsnotify.Chmod
}

type PatternsType []string
type BaseDir string

func SkipOp(ops ...fsnotify.Op) (ret fsnotify.Op) {
	for _, op := range ops {
		ret |= op
	}
	return ret
}

func Patterns(patterns ...string) PatternsType {
	return PatternsType(patterns)
}

func Watch(patterns PatternsType, args ...interface{}) {
	if len(patterns) > 0 {
		if len(args) == 0 {
			panic("watch must have related tasks or related function")
		}

		var err error
		w := &watch{}

		for i, arg := range args {
			switch value := arg.(type) {
			default:
				panic(fmt.Errorf("unknown args at arg[%d].", i+1))
			case func(event Event) error:
				w.fun = value
			case TasksType:
				w.tasks = []string(value)
			case BaseDir:
				w.baseDir = string(value)
			case fsnotify.Op:
				w.skipOp = value
			}
		}

		if len(w.tasks) == 0 && w.fun == nil {
			panic("watch must have related tasks or related function")
		}

		if err = w.initWatcher(patterns); err != nil {
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

		st := make(chan os.Signal)
		signal.Notify(st, os.Interrupt, os.Kill)
		<-st

		Log.Debug("Get shutdown signal.")
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

func (w *watch) initWatcher(patterns PatternsType) (err error) {
	if len(w.baseDir) == 0 {
		w.baseDir = "."
	}
	w.exit = make(chan bool)

	if w.skipOp == 0 {
		w.skipOp = SkipOp(Op_Chmod)
	}

	w.pselector, err = selector.New([]string(patterns)...)
	if err != nil {
		return fmt.Errorf("parse pattern err: %v", err)
	}

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
	if err = w.addAllDir(); err != nil {
		return err
	}

	// if err = w.addAllMatches(); err != nil {
	// 	return err
	// }

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
	Log.Debug("get event: %v", event)

	if event.Op&w.skipOp > 0 {
		Log.Debug("Event is skiped: %v", event)
		return nil
	}

	var bIsDir = false
	if fi, err := os.Stat(event.Name); err != nil {
		if !(event.IsRemove() && os.IsNotExist(err)) {
			Log.Warn("Get stat err: %v", err)
			return err
		}
	} else {
		if fi.IsDir() {
			if event.IsCreate() {
				err = w.watcher.Add(event.Name)
				Log.Debug("Add path '%s' to watcher. Error: %v", event.Name, err)
			}
		}

		bIsDir = fi.IsDir()
	}

	var bmatch bool
	if bmatch, err = w.pselector.Match(w.baseDir, event.Name, bIsDir); err != nil {
		return err
	}

	if bmatch {
		// if !bIsDir && event.IsCreate() {
		// 	w.watcher.Add(event.Name)
		// }
		w.doTask(event)
	}

	return nil
}

func (w *watch) doTask(event Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

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

func (w *watch) addAllMatches() (err error) {
	var matches []string
	if matches, err = w.pselector.Matches(w.baseDir); err != nil {
		return err
	}

	for _, match := range matches {
		if err = w.watcher.Add(match); err != nil {
			return err
		}
	}
	return nil
}
