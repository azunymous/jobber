package watch

import (
	v1 "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	testcore "k8s.io/client-go/testing"
	"k8s.io/utils/pointer"
	"time"
)

type FakeClient struct {
	*fake.Clientset
	pod     *core.Pod
	job     *v1.Job
	watcher *watch.FakeWatcher
}

func NewFakeClient() *FakeClient {
	pod := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      "a-pod",
			Namespace: "a-namespace",
			Labels:    map[string]string{"job-name": "a-job"},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "mainContainer",
					Image: "image",
				},
			},
		},
	}
	job := &v1.Job{
		ObjectMeta: meta.ObjectMeta{
			Name:      "a-job",
			Namespace: "a-namespace",
		},
		Spec: v1.JobSpec{
			Completions: pointer.Int32Ptr(1),
			Template:    core.PodTemplateSpec{Spec: pod.Spec},
		},
		Status: v1.JobStatus{
			Active:    1,
			Succeeded: 0,
			Failed:    0,
		},
	}

	fc := &FakeClient{Clientset: fake.NewSimpleClientset(pod, job)}
	fc.pod = pod
	fc.job = job
	watcher := watch.NewFake()
	fc.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
	fc.PrependWatchReactor("jobs", testcore.DefaultWatchReactor(watcher, nil))
	fc.watcher = watcher
	return fc
}

func (f *FakeClient) SendCompleteEvent(exitCode int32, in time.Duration) {
	time.AfterFunc(in, func() {
		f.pod.Status.ContainerStatuses = []core.ContainerStatus{
			{
				Name: "mainContainer",
				State: core.ContainerState{Terminated: &core.ContainerStateTerminated{
					ExitCode: exitCode,
					Signal:   0,
				}},
				Ready: true,
			},
		}
		f.watcher.Modify(f.pod)
	})
}

func (f *FakeClient) SendCompleteJobEvent(success bool, in time.Duration) {
	time.AfterFunc(in, func() {
		if success {
			f.job.Status.Succeeded = 1
		} else {
			f.job.Status.Failed = 1
		}
		f.watcher.Modify(f.job)
	})
}
