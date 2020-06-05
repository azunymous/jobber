package options

import (
	"github.com/spf13/cobra"
)

// Container struct contains options regarding the application
type Global struct {
	Verbosity int

}

func AddVerbosityArg(cmd *cobra.Command, g *Global) {
	cmd.PersistentFlags().IntVarP(&g.Verbosity, "verbosity", "v", 0, "Verbosity level")
}
