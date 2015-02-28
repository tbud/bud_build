package asset

import (
	. "github.com/tbud/bud/context"
	. "github.com/tbud/x/config"
	"testing"
)

func TestAsset(t *testing.T) {
	TaskConfig("bud.asset", Config{
		"includes": []string{"*.go"},
		"output":   "testdata/assets.go",
		"package":  "testdata",
	})
	// TaskConfig("asset.tobin.baseDir", "/Users/mind/gogo/src/github.com/tbud/x")

	// time.Sleep(30 * time.Second)
	UseTasks("bud")

	err := RunTask("asset")
	if err != nil {
		t.Error(err)
	}
}
