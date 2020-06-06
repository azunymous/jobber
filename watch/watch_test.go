package watch

import (
	"go.uber.org/zap"
	"jobber/test/check"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	testcore "k8s.io/client-go/testing"
	"testing"
	"time"
)

func TestNewContainerWatcher(t *testing.T) {
	fakeCS := fake.NewSimpleClientset()
	w := NewContainerWatcher(fakeCS, zap.L())

	check.Equals(t, &containerWatcher{
		clientset: fakeCS,
		logger:    zap.L(),
	}, w)
}

func Test_containerWatcher_Watch(t *testing.T) {
	type args struct {
		name      string
		pod       string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		pretest func(fc *FakeClient)
		wantErr bool
	}{
		{
			name: "watcher exits when container exits",
			args: args{
				name:      "mainContainer",
				pod:       "a-pod",
				namespace: "a-namespace",
			},
			pretest: func(fc *FakeClient) {
				fc.SendCompleteEvent(0, 10*time.Millisecond)
			},
			wantErr: false,
		},
		{
			name: "watcher exits when container exits on error",
			args: args{
				name:      "mainContainer",
				pod:       "a-pod",
				namespace: "a-namespace",
			},
			pretest: func(fc *FakeClient) {
				fc.SendCompleteEvent(1, 10*time.Millisecond)
			},
			wantErr: false,
		},
		{
			name: "watcher errors when container not found",
			args: args{
				name:      "non-existent container",
				pod:       "a-pod",
				namespace: "a-namespace",
			},
			pretest: func(fc *FakeClient) {},
			wantErr: true,
		},
		{
			name: "watcher errors when namespace not found",
			args: args{
				name:      "mainContainer",
				pod:       "a-pod",
				namespace: "non-existent namespace",
			},
			pretest: func(fc *FakeClient) {},
			wantErr: true,
		},
		{
			name: "watcher errors when pod not found",
			args: args{
				name:      "mainContainer",
				pod:       "nonexistent pod",
				namespace: "a-namespace",
			},
			pretest: func(fc *FakeClient) {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFakeClient()
			c := &containerWatcher{
				clientset: fc,
				logger:    zap.L(),
			}
			tt.pretest(fc)
			if err := c.Watch(tt.args.name, tt.args.pod, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("Watch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type FakeClient struct {
	*fake.Clientset
	pod     *core.Pod
	watcher *watch.FakeWatcher
}

func NewFakeClient() *FakeClient {
	pod := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name:      "a-pod",
			Namespace: "a-namespace",
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
	fc := &FakeClient{Clientset: fake.NewSimpleClientset(pod)}
	fc.pod = pod
	watcher := watch.NewFake()
	fc.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
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
