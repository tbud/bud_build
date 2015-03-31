// Copyright (c) 2015, tbud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import (
	. "github.com/tbud/bud/builtin"
	. "github.com/tbud/bud/context"
	_ "github.com/tbud/bud/tasks/asset"
	_ "github.com/tbud/bud/tasks/dist"
	_ "github.com/tbud/bud/tasks/license"
	"os"
)

func init() {
	// bud clean task
	Task("clean", BUD_TASK_GROUP, Usage("Clean bud script run temp dir."), func() error {
		return FindFiles(".budtmp.*").Each(os.RemoveAll)
	})
}
