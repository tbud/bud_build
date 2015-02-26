package asset

import (
	. "github.com/tbud/bud/context"
	. "github.com/tbud/x/config"
	"testing"
)

func TestAsset(t *testing.T) {
	TaskConfig("asset.tobin", Config{
		"includes": []string{"*.go"},
		"output":   "testdata/assets.go",
		"package":  "testdata",
	})
	// TaskConfig("asset.tobin.baseDir", "/Users/mind/gogo/src/github.com/tbud/x")

	UseTasks()

	err := RunTask("tobin")
	if err != nil {
		t.Error(err)
	}
}
