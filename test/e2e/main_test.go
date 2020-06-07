//+build e2e

package e2e

import (
	"fmt"
	"github.com/prometheus/common/log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	Jobber  = "./jobber"
	Context string
)

func init() {
	// Move to project root
	must(os.Chdir("../.."))
	cmd, set := os.LookupEnv("JOBBER_CMD")
	if set {
		fmt.Println("Using command " + cmd)
		Jobber = cmd
	}
	context := os.Getenv("JOBBER_TEST_CONTEXT")
	if context == "" {
		context = "kind-jobber"
		createKindCluster()
	}

	/*
		Run skaffold with either kind context or other context
		Run jobber wait on Job
		Check bucket contains required items
	*/
}

func createKindCluster() {
	b, err := exec.Command("kind", "get", "clusters").CombinedOutput()
	must(err)
	if strings.Contains(string(b), "jobber") {
		log.Info("Using existing cluster with context 'kind-jobber'")
		return
	}
	RunCmd("kind", "delete", "cluster", "--name", "jobber")
	RunCmd("kind", "create", "cluster", "--name", "jobber", "--config", "kind.yaml")
	log.Info("Waiting 60s for nodes to be ready")
	time.Sleep(60 * time.Second)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func RunCmd(c ...string) {
	cmd := exec.Command(c[0], c[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}
