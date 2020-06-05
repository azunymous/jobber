package options

import (
	"github.com/spf13/cobra"
)

// Container struct contains options regarding the application
type Container struct {
	Name string

}

func AddNameArg(cmd *cobra.Command, c *Container) {
	cmd.PersistentFlags().StringVarP(&c.Name, "name", "n", "", "Container name to monitor")
	must(cmd.MarkPersistentFlagRequired("name"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
