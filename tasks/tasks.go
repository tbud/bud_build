package tasks

import (
	. "github.com/tbud/bud/builtin"
	. "github.com/tbud/bud/context"
	_ "github.com/tbud/bud/tasks/asset"
	"os"
)

func init() {
	// bud clean task
	Task("clean", BUD_TASK_GROUP, Usage("clean bud script run temp dir."), func() error {
		return FindFiles(".budtmp.*").Each(os.RemoveAll)
	})
}
