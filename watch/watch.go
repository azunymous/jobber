package watch

import (
	"fmt"
	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

type containerWatcher struct {
	clientset kubernetes.Interface
	logger    *zap.Logger
}

func NewContainerWatcher(clientset kubernetes.Interface, logger *zap.Logger) *containerWatcher {
	return &containerWatcher{clientset: clientset, logger: logger}
}

func (c *containerWatcher) Watch(name, pod, namespace string) error {
	c.logger = c.logger.
		With(zap.String("container", name), zap.String("pod", pod), zap.String("namespace", namespace))
	c.logger.Info("Starting watch")

	c.logger.Debug("Getting pod resource")
	p, err := c.clientset.CoreV1().Pods(namespace).Get(pod, meta.GetOptions{})
	if err != nil {
		return err
	}

	c.logger.Debug("Checking container exists")
	err = getContainer(name, p.Spec.Containers)
	if err != nil {
		return err
	}

	c.logger.Debug("Watching pod events")
	watch, err := c.clientset.CoreV1().Pods(namespace).
		Watch(meta.ListOptions{Watch: true, FieldSelector: fields.Set{"metadata.name": pod}.AsSelector().String()})
	if err != nil {
		return err
	}
	defer watch.Stop()
	for event := range watch.ResultChan() {
		c.logger.Info("Event received", zap.String("event", string(event.Type)))
		p := event.Object.(*core.Pod)
		status, terminated, err := checkIfTerminated(name, p.Status.ContainerStatuses)
		if err != nil {
			return err
		}

		if terminated {
			c.logger.Info("Monitored container terminated", zap.Int32("exitCode", status.State.Terminated.ExitCode))
			return nil
		}
	}
	return nil
}

func getContainer(name string, containers []core.Container) error {
	for _, container := range containers {
		if container.Name == name {
			return nil
		}
	}
	return fmt.Errorf("unable to find container %s in containers", name)
}

func checkIfTerminated(name string, statuses []core.ContainerStatus) (*core.ContainerStatus, bool, error) {
	for _, status := range statuses {
		if status.Name == name {
			if status.State.Terminated != nil {
				return &status, true, nil
			}
			return &status, false, nil
		}
	}
	return nil, false, fmt.Errorf("unable to find container %s in containers", name)
}
