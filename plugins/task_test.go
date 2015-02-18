package plugins

import (
	"fmt"
	"reflect"
	"testing"
)

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
