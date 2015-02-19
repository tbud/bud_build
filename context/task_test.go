package context

import (
	"fmt"
	"reflect"
	"testing"
)

func init() {
	Task("A1", func() error {
		return nil
	})

	Task("B1", Depend("A1"), func() error {
		return nil
	})

	TaskPackageToDefault()
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
