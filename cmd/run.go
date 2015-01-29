package cmd

import (
	"github.com/tbud/bud/seed"
)

var cmdRun = &Command{
	Run:       runCommand,
	UsageLine: "run [import path] [run mode] [port]",
	Short:     "run a application",
	Long: `
Run the bud web application named by the given import path.

For example, to run the chat room sample application:

    bud run github.com/tbud/samples/chat dev

The run mode is used to select which set of app.conf configuration should
apply and may be used to determine logic in the application itself.

Run mode defaults to "dev".

You can set a port as an optional third parameter.  For example:

    bud run github.com/tbud/samples/chat prod 8080
    `,
}

func runCommand(cmd *Command, args []string) {
	if len(args) == 0 {
		fatalf("No import path given.\nRun 'bud help run' for usage.\n")
	}

	emb := seed.NewEmbryo()

	emb.Run()
}
