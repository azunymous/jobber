package commands

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"jobber/commands/options"
	"jobber/uploader"
	"jobber/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"time"
)

// addRelease adds the increment command to a top level command.
func addMonitor(topLevel *cobra.Command) {
	globalOpts := &options.Global{}
	containerOpts := &options.Container{}
	monitor := &cobra.Command{
		Use:   "monitor",
		Short: "Start monitoring a container",
		Long: `Monitor a container. This should be run as a sidecar in a Job.

This command expects the environment variables 'POD_NAME' and 'NAMESPACE_NAME' to be set via 
the Kubernetes downward API. A service account token with the ability to read pods in this namespace
is also required.

This is so it can monitor the state of the main container in the pod.

(Note: This access is required until Kubernetes supports the sidecar container lifecycle)
`,
		Run: func(cmd *cobra.Command, args []string) {
			logger, _ := zap.NewProduction()
			if globalOpts.Verbosity != 0 {
				logger, _ = zap.NewDevelopment()
			}
			err := monitor(logger, containerOpts)
			logger.Info("Monitor complete")
			if err != nil {
				logger.Fatal("Monitor failed", zap.Error(err))
			}
		},
	}
	options.AddVerbosityArg(monitor, globalOpts)
	options.AddNameArg(monitor, containerOpts)
	options.AddCopyFolderArg(monitor, containerOpts)
	topLevel.AddCommand(monitor)
}

func monitor(logger *zap.Logger, containerOpts *options.Container) error {
	logger.Info("Starting monitor of job container " + containerOpts.Name)
	podName := os.Getenv("POD_NAME")
	namespaceName := os.Getenv("NAMESPACE_NAME")

	c, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset := kubernetes.NewForConfigOrDie(c)
	w := watch.NewContainerWatcher(clientset, logger)
	err = w.Watch(containerOpts.Name, podName, namespaceName)
	if err != nil {
		return fmt.Errorf("watch error: %v", err)
	}

	if len(containerOpts.UploadFile) > 0 {
		logger.Sugar().Debugf("Starting file uploads %v", containerOpts.UploadFile)
		u := uploader.NewUploaderOrDie(uploader.Config{
			Endpoint:  os.Getenv("JOBBER_ENDPOINT"),
			AccessKey: os.Getenv("JOBBER_ACCESS_KEY"),
			SecretKey: os.Getenv("JOBBER_SECRET_KEY"),
		}, logger)
		err = u.Initialize(containerOpts.Name)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		g, ctx := errgroup.WithContext(ctx)
		for _, f := range containerOpts.UploadFile {
			// Pass the current value into the goroutine closure rather than the variable
			f := f
			g.Go(func() error {
				logger.Info("Uploading file", zap.String("file", f))
				return u.Upload(containerOpts.Name, podName, f)
			})
		}
		return g.Wait()
	}
	return nil
}
