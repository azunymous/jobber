package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"jobber/commands/options"
	"jobber/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

// addRelease adds the increment command to a top level command.
func addWait(topLevel *cobra.Command) {
	globalOpts := &options.Global{}
	kubectlOpts := &options.Kubectl{}
	wait := &cobra.Command{
		Use:   "wait [Job Name]",
		Short: "Wait for a Job",
		Long: `Wait for a Job to complete
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logger, _ := zap.NewProduction()
			if globalOpts.Verbosity != 0 {
				logger, _ = zap.NewDevelopment()
			}
			err := wait(logger, args[0], kubectlOpts)
			logger.Info("Wait complete")
			if err != nil {
				logger.Fatal("Wait failed", zap.Error(err))
			}
		},
	}
	options.AddContextArg(wait, kubectlOpts)
	options.AddNamespaceArg(wait, kubectlOpts)
	options.AddVerbosityArg(wait, globalOpts)
	topLevel.AddCommand(wait)
}

func wait(logger *zap.Logger, name string, opts *options.Kubectl) error {
	logger.Info("Starting wait of job " + name)

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	if opts.Context != "" {
		logger = logger.With(zap.String("context", opts.Context))
		logger.Info("Using provided context")
		configOverrides.CurrentContext = opts.Context
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	c, err := kubeConfig.ClientConfig()
	if err != nil {
		return err
	}
	clientset := kubernetes.NewForConfigOrDie(c)
	w := watch.NewContainerWatcher(clientset, logger)
	err = w.Wait(os.Stdout, name, opts.Namespace)
	if err != nil {
		return fmt.Errorf("watch error: %v", err)
	}
	return nil
}
