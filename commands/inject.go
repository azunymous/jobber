package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"io/ioutil"
	"jobber/commands/options"
	"jobber/parser"
	"os"
)

// addRelease adds the increment command to a top level command.
func addInject(topLevel *cobra.Command) {
	globalOpts := &options.Global{}
	inject := &cobra.Command{
		Use:   "inject [file]",
		Short: "[PRERELEASE] Inject jobber into a Job",
		Long: `Inject jobber into a Job as a sidecar

Currently a prototype.
This currently uses the first container in a Job pod spec as the one that is monitored and no resources are defined for
the jobber container. It also does not configure any file uploading. It is currently far easier to configure your
Job YAML directly instead of using the inject command.
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger, _ := zap.NewProduction()
			if globalOpts.Verbosity != 0 {
				logger, _ = zap.NewDevelopment()
			}
			err := inject(logger, args[0])
			if err != nil {
				logger.Sugar().Fatal("Injection failed", zap.Error(err))
			}
		},
	}
	options.AddVerbosityArg(inject, globalOpts)
	topLevel.AddCommand(inject)
}

func inject(logger *zap.Logger, file string) error {
	in, err := input(file)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", file, err)
	}
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	parser.InjectJobs(os.Stdout, b)
	return nil
}

func input(file string) (*os.File, error) {
	if file == "-" {
		return os.Stdin, nil
	}
	return os.Open(file)
}
