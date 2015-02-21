package context

import (
	"fmt"
	"github.com/tbud/x/config"
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
}

func (t *testTask) Execute() error {
	fmt.Println(t)
	return nil
}

func (t *testTask) Validate() error {
	fmt.Println("test task validate")
	return nil
}

func init() {
	Task("A1", func() error {
		return nil
	})

	Task("B1", Depend("A1"), func() error {
		return nil
	})

	TaskPackageToDefault()
}

func TestTestTask(t *testing.T) {
	Task("test", &testTask{}, config.Config{
		"package":  "pkg",
		"baseDir":  "bb",
		"compress": true,
		"num":      3,
		"test":     1,
		"includes": []string{"resource/**/*.go", "temp/**/*.tmpl"},
		"excludes": []string{"1", "2", "3"},
	})
	TaskPackageToDefault()

	RunTask("test")
}

func BenchmarkTaskRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunTask("B1")
	}
}

func TestDepends(t *testing.T) {
	dep := Depend("a", "b", "c")
	got := []string{"a", "b", "c"}
	if !reflect.DeepEqual(dep, Depends(got)) {
		t.Errorf("want %v, got %v", dep, got)
	}
}

func TestTask(t *testing.T) {
	Task("A", func() error {
		fmt.Println("in A")
		return nil
	})

	Task("B", Depend("A"), func() error {
		fmt.Println("in B")
		return nil
	})

	TaskPackageToDefault()

	err := RunTask("B")
	if err != nil {
		t.Error(err)
	}
}
