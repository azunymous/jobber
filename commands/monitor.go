package commands

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"jobber/commands/options"
	"jobber/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

// addRelease adds the increment command to a top level command.
func addMonitor(topLevel *cobra.Command) {
	globalOpts := &options.Global{}
	containerOpts := &options.Container{}
	monitor := &cobra.Command{
		Use:   "monitor",
		Short: "Start monitoring a container",
		Long: `Monitor container. This should be run as a sidecar
`,
		Run: func(cmd *cobra.Command, args []string) {
			logger, _ := zap.NewProduction()
			if globalOpts.Verbosity != 0 {
				logger, _ = zap.NewDevelopment()
			}
			err := monitor(logger, containerOpts)
			logger.Info("Monitor complete")
			if err != nil {
				logger.Fatal("Failed", zap.Error(err))
			}
		},
	}
	options.AddVerbosityArg(monitor, globalOpts)
	options.AddNameArg(monitor, containerOpts)
	topLevel.AddCommand(monitor)
}

func monitor(logger *zap.Logger, containerOpts *options.Container) error {
	logger.Info("Starting monitor of job container " + containerOpts.Name)
	c, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset := kubernetes.NewForConfigOrDie(c)
	w := watch.NewContainerWatcher(clientset, logger)
	return w.Watch(containerOpts.Name, os.Getenv("POD_NAME"), os.Getenv("NAMESPACE_NAME"))
}
