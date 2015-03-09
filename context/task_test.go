package context

import (
	"fmt"
	. "github.com/tbud/x/config"
	"reflect"
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
	fmt.Println(t)
	return nil
}

func (t *testTask) Validate() error {
	if t.Num != 3 || t.Package != "pkg" || t.BaseDir != "bb" || t.Float1 != 1.18 ||
		!reflect.DeepEqual(t.Excludes, []string{"1", "2", "3"}) ||
		!reflect.DeepEqual(t.Includes, []string{"resource/**/*.go", "temp/**/*.tmpl"}) {
		return fmt.Errorf("there are task config set error")
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
	TaskConfig("", Config{
		CONTEXT_CONFIG_TASK_KEY: Config{
			"context": Config{
				"test": Config{
					"package":  "pkg",
					"num":      4,
					"test":     1,
					"float1":   1.18,
					"includes": []string{"resource/**/*.go"},
					"excludes": []string{"1", "2", "3"},
				},
			},
		},
	})

	Task("test", &testTask{}, Config{
		"baseDir":  "bb",
		"compress": true,
		"num":      3,
		"includes": []string{"resource/**/*.go", "temp/**/*.tmpl"},
	})
	UseTasks()

	RunTask("test")
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

func TestFullTask(t *testing.T) {
	Task("f1", Group("py"), Usage("for test"), func() error {
		fmt.Println("in f1")
		return nil
	})

	RunTask("py.f1")
}
