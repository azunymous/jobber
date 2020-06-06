package options

import (
	"github.com/spf13/cobra"
)

// Container struct contains options regarding the application
type Container struct {
	Name       string
	UploadFile []string
}

func AddNameArg(cmd *cobra.Command, c *Container) {
	cmd.PersistentFlags().StringVarP(&c.Name, "name", "n", "", "Container name to monitor")
	must(cmd.MarkPersistentFlagRequired("name"))
}

func AddCopyFolderArg(cmd *cobra.Command, c *Container) {
	cmd.Flags().StringSliceVarP(&c.UploadFile, "upload-file", "u", nil, "Files to be copied")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
