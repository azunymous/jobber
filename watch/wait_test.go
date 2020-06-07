package watch

import (
	"bytes"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"io"
	"io/ioutil"
	"jobber/test/check"
	"testing"
	"time"
)

// TODO This cannot be tested with the fake client until https://github.com/kubernetes/kubernetes/pull/91485 is released

func DisabledTest_containerWatcher_Wait(t *testing.T) {
	type args struct {
		w         io.Writer
		job       string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		pretest func(fc *FakeClient)
		wantErr bool
	}{
		{
			name: "watches to completion",
			args: args{
				w:         &buffer.Buffer{},
				job:       "a-job",
				namespace: "a-namespace",
			},
			pretest: func(fc *FakeClient) {
				fc.SendCompleteJobEvent(true, 10*time.Millisecond)
			},
			wantErr: false,
		},
		{
			name: "watches to completion on error",
			args: args{
				w:         &buffer.Buffer{},
				job:       "a-job",
				namespace: "a-namespace",
			},
			pretest: func(fc *FakeClient) {
				fc.SendCompleteJobEvent(false, 10*time.Millisecond)
			},
			wantErr: false,
		},
		{
			name: "errors if job doesn't exist",
			args: args{
				w:         &buffer.Buffer{},
				job:       "non-existent-job",
				namespace: "a-namespace",
			},
			pretest: func(fc *FakeClient) {},
			wantErr: true,
		},
		{
			name: "errors if namespace doesn't exist",
			args: args{
				w:         &buffer.Buffer{},
				job:       "a-job",
				namespace: "non-existent-namespace",
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
			if err := c.Wait(tt.args.w, tt.args.job, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("Wait() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func DisabledTestWaitWritesLogs(t *testing.T) {
	fc := NewFakeClient()
	c := &containerWatcher{
		clientset: fc,
		logger:    zap.L(),
	}
	fc.SendCompleteJobEvent(true, 10*time.Millisecond)

	var b bytes.Buffer
	err := c.Wait(&b, "a-job", "a-namespace")

	check.Ok(t, err)
	all, err := ioutil.ReadAll(&b)
	check.Ok(t, err)

	check.Equals(t, "fake logs", string(all))
}
