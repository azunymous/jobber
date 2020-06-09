package watch

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"io"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// Wait waits for a job to finish; printing it's logs as it waits
func (c *containerWatcher) Wait(w io.Writer, job, namespace string) error {
	c.logger = c.logger.With(zap.String("job", job), zap.String("namespace", namespace))
	c.logger.Info("Waiting for job")
	err := c.retry(5, 1*time.Second, func() error {
		_, err := c.clientset.BatchV1().Jobs(namespace).Get(job, meta.GetOptions{})
		return err
	})
	if err != nil {
		return err
	}

	c.logger.Info("Found job")
	var pods *v1.PodList
	err = c.retry(10, 1*time.Second, func() error {
		var err error
		pods, err = c.clientset.CoreV1().Pods(namespace).List(meta.ListOptions{LabelSelector: "job-name=" + job})
		if err != nil {
			return err
		}
		if len(pods.Items) < 1 {
			return errors.New("no pods found for job")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(pods.Items) < 1 {
		return errors.New("no pods found for job")
	}
	p := pods.Items[0]
	cName := getFirstContainer(p).Name

	c.logger.Info("Waiting for pod to be ready")
	err = c.retry(10, 1*time.Second, func() error {
		checkedPod, err := c.clientset.CoreV1().Pods(namespace).Get(p.Name, meta.GetOptions{})
		if err != nil {
			return err
		}
		if len(checkedPod.Status.ContainerStatuses) < 0 {
			return fmt.Errorf("no container status available in pod %s", checkedPod.Name)
		}
		status := getContainerStatus(cName, checkedPod.Status.ContainerStatuses)
		if status == nil || !status.Ready {
			return fmt.Errorf("container %s is not ready", cName)
		}
		return nil
	})

	if err != nil {
		c.logger.Warn("Cannot determine if pod is ready. Continuing anyway.")
	}

	c.logger.Info("Streaming logs")
	stream, err := c.clientset.CoreV1().Pods(namespace).GetLogs(p.Name, &v1.PodLogOptions{Container: cName, Follow: true}).Stream()
	if err != nil {
		return err
	}
	defer stream.Close()

	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		scanner := bufio.NewScanner(stream)
		for {
			select {
			case <-ctx.Done():
				time.Sleep(1 * time.Second)
				return nil
			default:
				scanned := scanner.Scan()
				if !scanned {
					return scanner.Err()
				}
				_, err := fmt.Fprintln(w, scanner.Text())
				if err != nil {
					return err
				}
			}
		}
	})

	g.Go(func() error {
		select {
		case <-ctx.Done():
			time.Sleep(1 * time.Second)
			return nil
		default:
			for {
				j, err := c.clientset.BatchV1().Jobs(namespace).Get(job, meta.GetOptions{})
				if err != nil {
					return err
				}
				if j.Status.Failed > 0 {
					return errors.New("job failed")
				}
				if j.Status.Succeeded > 0 {
					return nil
				}
			}
		}
	})

	return g.Wait()
}

func getContainerStatus(name string, statuses []v1.ContainerStatus) *v1.ContainerStatus {
	for _, status := range statuses {
		if status.Name == name {
			return &status
		}
	}
	return nil
}

func getFirstContainer(p v1.Pod) v1.Container {
	if len(p.Spec.Containers) < 1 {
		panic("No containers in Job pod!")
	}
	return p.Spec.Containers[0]
}

func (c *containerWatcher) retry(attempts int, sleep time.Duration, f func() error) error {
	var err error
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return nil
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)

		c.logger.Warn("retrying after error", zap.Error(err))
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
