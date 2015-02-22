package asset

import (
	. "github.com/tbud/bud/context"
	"testing"
)

func TestAsset(t *testing.T) {
	TaskPackageToDefault()

	err := RunTask("tobin")
	if err != nil {
		t.Error(err)
	}
}
