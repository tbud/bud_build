package script

import (
	"path/filepath"
	"testing"
)

func TestScript(t *testing.T) {
	file, err := filepath.Abs("script_test.bud")
	if err != nil {
		t.Error(err)
	}

	err = Run(file)
	if err != nil {
		t.Error(err)
	}
}
