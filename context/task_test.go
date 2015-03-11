// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	. "github.com/tbud/x/config"
	"reflect"
	"strings"
	"testing"
)

type testTask struct {
	Package  string
	BaseDir  string
	Includes []string
	Excludes []string
	Output   string
	Compress bool
	Num      int
	Float1   float64
}

func (t *testTask) Execute() error {
	fmt.Printf("Ok")
	return nil
}

func (t *testTask) Validate() error {
	if t.Num != 4 || t.Package != "pkg" || t.BaseDir != "dd" || t.Float1 != 1980.01 ||
		t.Compress != true ||
		!reflect.DeepEqual(t.Excludes, []string{"1", "2", "3"}) ||
		!reflect.DeepEqual(t.Includes, []string{"resource/**/*.go"}) {
		return fmt.Errorf("there are task config set error.%#v\n", t)
	}
	return nil
}

func init() {
	Task("A1", func() error {
		return nil
	})

	Task("B1", Tasks("A1"), func() error {
		return nil
	})

	UseTasks()
}

func TestTestTask(t *testing.T) {
	Task("test", &testTask{}, Group("tasktest"), Config{
		"baseDir":  "bb",
		"compress": true,
		"num":      3,
		"includes": []string{"resource/**/*.go", "temp/**/*.tmpl"},
	})

	TaskConfig("tasktest.test", Config{
		"package":  "pkg",
		"num":      4,
		"test":     1,
		"float1":   1.18,
		"includes": []string{"resource/**/*.go"},
		"excludes": []string{"1", "2", "3"},
	})

	TaskConfig("tasktest", Config{
		"baseDir": "dd",
		"num":     5,
	})

	UseTasks("tasktest")

	if err := RunTask("test", Config{"float1": 1980.01}); err != nil {
		t.Error(err)
	}
}

func BenchmarkTaskRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunTask("B1")
	}
}

func TestTasks(t *testing.T) {
	dep := Tasks("a", "b", "c")
	got := []string{"a", "b", "c"}
	if !reflect.DeepEqual(dep, TasksType(got)) {
		t.Errorf("want %v, got %v", dep, got)
	}
}

func TestTask(t *testing.T) {
	Task("A", func() error {
		fmt.Println("in A")
		return nil
	})

	Task("B", Tasks("A"), func() error {
		fmt.Println("in B")
		return nil
	})

	UseTasks()

	err := RunTask("B")
	if err != nil {
		t.Error(err)
	}
}

func TestTaskRecursion(t *testing.T) {
	Task("A2", Tasks("C2"), func() error {
		fmt.Println("in A")
		return nil
	})

	Task("B2", Tasks("A2"), func() error {
		fmt.Println("in B")
		return nil
	})

	Task("C2", Tasks("B2"), func() error {
		fmt.Println("in B")
		return nil
	})

	UseTasks()

	err := RunTask("C2")
	if err == nil || !strings.Contains(err.Error(), "recursion call") {
		t.Error(err)
	}
}

func TestFullTask(t *testing.T) {
	Task("f1", Group("py"), Usage("for test"), func() error {
		fmt.Println("in f1")
		return nil
	})

	RunTask("py.f1")
}
