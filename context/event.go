package context

import (
	"gopkg.in/fsnotify.v1"
)

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
