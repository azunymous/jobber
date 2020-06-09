//+build e2e

package e2e

import (
	"github.com/minio/minio-go/v6"
	"io/ioutil"
	"jobber/test/check"
	"os/exec"
	"strings"
	"testing"
)

func TestJobberLogs(t *testing.T) {
	RunCmd("skaffold", "run", "--kube-context", Context, "-p", "e2e")
	t.Cleanup(stop("e2e"))
	out, err := exec.Command(Jobber, "wait", "testjob", "--namespace", "default", "--context", Context).CombinedOutput()
	output := string(out)
	t.Log("---Output---")
	t.Log(output)
	t.Log("---Output End---")
	check.Ok(t, err)
	check.Assert(t, strings.Contains(output, "expected-log-message"), "expected output to contain 'expected-log-message'")
}

func TestJobberLogsFailure(t *testing.T) {
	RunCmd("skaffold", "run", "--kube-context", Context, "-p", "e2e-fail")
	t.Cleanup(stop("e2e-fail"))
	command := exec.Command(Jobber, "wait", "testjob", "--namespace", "default", "--context", Context)
	out, err := command.CombinedOutput()
	output := string(out)
	t.Log("---Output---")
	t.Log(output)
	t.Log("---Output End---")
	check.Equals(t, 1, command.ProcessState.ExitCode())
	check.Assert(t, err != nil, "Expected error, but didn't get an error")
	check.Assert(t, strings.Contains(output, "expected-log-message"), "expected output to contain 'expected-log-message'")
}

func TestJobberUploads(t *testing.T) {
	RunCmd("skaffold", "run", "--kube-context", Context, "-p", "e2e")
	t.Cleanup(stop("e2e"))
	b, err := exec.Command(Jobber, "wait", "testjob", "--namespace", "default", "--context", Context).CombinedOutput()
	t.Log(string(b))
	check.Ok(t, err)
	mc, err := minio.New("localhost:9000", "minio", "insecure", false)
	check.Ok(t, err)
	doneCh := make(chan struct{})
	defer close(doneCh)
	objs := mc.ListObjectsV2("testjob", "", true, doneCh)
	obj := <-objs
	// Object metadata also contains any error from listing objects
	t.Log(obj)
	object, err := mc.GetObject("testjob", obj.Key, minio.GetObjectOptions{})
	check.Ok(t, err)
	defer object.Close()
	content, err := ioutil.ReadAll(object)
	check.Ok(t, err)

	check.Equals(t, "content\n", string(content))
}

func TestJobberFails(t *testing.T) {
	RunCmd("skaffold", "run", "--kube-context", Context, "-p", "e2e-fail")
	t.Cleanup(stop("e2e-fail"))
	command := exec.Command(Jobber, "wait", "testjob", "--namespace", "default", "--context", Context)
	out, err := command.CombinedOutput()
	t.Log("---Output---")
	t.Log(string(out))
	t.Log("---Output End---")
	check.Equals(t, 1, command.ProcessState.ExitCode())
	check.Assert(t, err != nil, "Expected error, but didn't get an error")
}

func TestJobberFailsUploads(t *testing.T) {
	RunCmd("skaffold", "run", "--kube-context", Context, "-p", "e2e-fail")
	t.Cleanup(stop("e2e-fail"))
	command := exec.Command(Jobber, "wait", "testjob", "--namespace", "default", "--context", Context)
	_, err := command.CombinedOutput()
	check.Equals(t, 1, command.ProcessState.ExitCode())
	mc, err := minio.New("localhost:9000", "minio", "insecure", false)
	check.Ok(t, err)
	doneCh := make(chan struct{})
	defer close(doneCh)
	objs := mc.ListObjectsV2("testjob", "", true, doneCh)
	obj := <-objs
	// Object metadata also contains any error from listing objects
	t.Log(obj)
	object, err := mc.GetObject("testjob", obj.Key, minio.GetObjectOptions{})
	check.Ok(t, err)
	defer object.Close()
	content, err := ioutil.ReadAll(object)
	check.Ok(t, err)

	check.Equals(t, "content\n", string(content))
}

func stop(profile string) func() {
	return func() {
		RunCmd("skaffold", "delete", "--kube-context", Context, "-p", profile)
	}
}
