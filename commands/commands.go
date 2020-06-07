// Commands package specifies all subcommands of the top level command
package commands

import (
	"github.com/spf13/cobra"
)

// Add commands all the commands to a top level command.
func AddCommands(topLevel *cobra.Command) {
	addMonitor(topLevel)
	addInject(topLevel)
	addWait(topLevel)
}
