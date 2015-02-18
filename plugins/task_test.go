package plugins

import (
	"fmt"
	"github.com/tbud/bud/context"
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
	Task("A", Depend("B"), func(context context.Context, args ...string) error {
		fmt.Println("in A")
		fmt.Println(args)
		return nil
	})

	Task("B", Depend("A"), func(context context.Context, args ...string) error {
		fmt.Println("in B")
		fmt.Println(args)
		return nil
	})

	// RunTask("A")
	RunTask("B")
}
