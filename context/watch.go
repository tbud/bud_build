// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	. "github.com/tbud/x/builtin"
	"github.com/tbud/x/path/selector"
	"gopkg.in/fsnotify.v1"
	"os"
	"os/signal"
	"sync"
	"time"
)

var _watchs []*watch
var stopWatch chan bool

type watch struct {
	// option param
	baseDir    string
	pselector  *selector.Selector
	tasks      []string                   // run task when event notify
	fun        func(events []Event) error // run when event notify
	skipOp     fsnotify.Op                // ops that will be skiped, default is Op_Chmod
	waitMsec   time.Duration              // if wait time is 0, event will send immediately; otherwise will wait x msec,default is 100
	mergeEvent bool                       // when wait msec is not 0, event will merge when path is same

	// use inside
	watcher    *fsnotify.Watcher // watcher that watch the dir
	exit       chan bool         // use to graceful stop watch
	events     []Event           // events need to delay notify
	eventsLock sync.Mutex        // lock for access the events
	taskLock   sync.Mutex        // lock for run task
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
		w := &watch{waitMsec: 100, mergeEvent: true}

		for i, arg := range args {
			switch value := arg.(type) {
			default:
				panic(fmt.Errorf("unknown args at arg[%d].", i+1))
			case func(events []Event) error:
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

	// comment add matches
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
	Log.Trace("get event: %v", event)

	if event.Op&w.skipOp > 0 {
		Log.Trace("Event is skiped: %v", event)
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
		// comment add matches
		// if !bIsDir && event.IsCreate() {
		// 	w.watcher.Add(event.Name)
		// }
		if w.waitMsec == 0 {
			if err = w.doTask([]Event{event}); err != nil {
				Log.Warn("do task error: %v", err)
				return err
			}
		} else {
			if err = w.delayDoTask(event); err != nil {
				Log.Warn("delay task error: %v", err)
				return err
			}
		}
	}

	return nil
}

func (w *watch) delayDoTask(event Event) error {
	w.eventsLock.Lock()
	defer w.eventsLock.Unlock()

	if len(w.events) == 0 {
		go func() {
			time.Sleep(w.waitMsec * time.Millisecond)
			var e []Event
			w.eventsLock.Lock()
			e = w.events
			w.events = w.events[:0]
			w.eventsLock.Unlock()
			w.doTask(e)
		}()
	}

	if w.mergeEvent {
		for i, evn := range w.events {
			if evn.Name == event.Name {
				w.events[i].Op |= event.Op
				return nil
			}
		}
	}

	w.events = append(w.events, event)
	return nil
}

func (w *watch) doTask(events []Event) (err error) {
	defer Catch(func(ierr interface{}) {
		switch value := ierr.(type) {
		case error:
			err = value
		}
		Log.Error("Catch error: %v", ierr)
	})

	w.taskLock.Lock()
	defer w.taskLock.Unlock()

	if len(w.tasks) > 0 {
		for _, task := range w.tasks {
			if err = RunTask(task); err != nil {
				return err
			}
		}
	}

	if w.fun != nil {
		if err = w.fun(events); err != nil {
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

// comment add matches
// func (w *watch) addAllMatches() (err error) {
// 	var matches []string
// 	if matches, err = w.pselector.Matches(w.baseDir); err != nil {
// 		return err
// 	}

// 	for _, match := range matches {
// 		if err = w.watcher.Add(match); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
