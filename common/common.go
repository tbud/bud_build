package common

import (
	"os"
	"sync"
)

var exitStatus = 0

var exitMu sync.Mutex

func SetExitStatus(n int) {
	exitMu.Lock()
	if exitStatus < n {
		exitStatus = n
	}
	exitMu.Unlock()
}

var atexitFuncs []func()

func atexit(f func()) {
	atexitFuncs = append(atexitFuncs, f)
}

func Exit() {
	for _, f := range atexitFuncs {
		f()
	}
	os.Exit(exitStatus)
}

func PanicIfError(err error) {
	if err != nil {
		Log.Error(err.Error())
		panic(err)
	}
}

func LogFatalExit(format string, args ...interface{}) {
	Log.Fatal(format, args...)
	os.Exit(1)
}

// func errorf(format string, args ...interface{}) {
// 	Log.Error(format, args...)
// 	SetExitStatus(1)
// }

// func exitIfErrors() {
// 	if exitStatus != 0 {
// 		Exit()
// 	}
// }
