package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"jobber/commands"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "jobber",
		Short:   "Kubernetes Job monitoring",
		Long:    `Monitoring Kubernetes Jobs such as for testing`,
		Version: version,
	}

	rootCmd.SetVersionTemplate(versionTemplate())
	commands.AddCommands(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func versionTemplate() string {
	return `{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
Git Hash: ` + gitHash + `
Time: ` + buildTime + ` UTC
`
}
