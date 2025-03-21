package google

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"

	"github.com/feight/deploy/tui"
)

type KubernetesTarget struct {
	GoogleTarget
}

func (t *KubernetesTarget) Text() string {
	return fmt.Sprintf("[%s, Kubernetes Engine]", t.ProjectId)
}

func (t *KubernetesTarget) setCluster() {

	cmd := exec.Command(
		"gcloud",
		"container",
		"clusters",
		"get-credentials",
		"workers",
		"--zone", t.Region,
		"--project", t.ProjectId)

	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete get-credentials"))
	}
}

func (t *KubernetesTarget) Deploy() {

	t.setCluster()
	t.Stop()

	cmd := exec.Command(
		"kubectl",
		"run",
		t.serviceName,
		"--image", t.GetImageTag(),
		"--image-pull-policy", "Always",
	)

	for _, v := range t.Environment {
		cmd.Args = append(cmd.Args, []string{"--env", v}...)
	}

	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete deploy"))
	}
}

func (t *KubernetesTarget) PostDeploy() {

	fmt.Printf("> Waiting for logs...\n\n")

	/* Wait 10s for the pod to start... */
	time.Sleep(time.Second * 10)

	t.TailLogs()
}

func (t *KubernetesTarget) TailLogs() {

	cmd := exec.Command(
		"kubectl",
		"logs", "-f",
		t.serviceName,
	)

	var buf bytes.Buffer

	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = &buf

	err := cmd.Run()

	if err != nil {

		fmt.Printf("Waiting for pod to start (%s)...\n\n",
			color.HiRedString("%s", strings.Trim(buf.String(), "\n")))

		time.Sleep(time.Second * 5)

		t.TailLogs()
	}
}

func (t *KubernetesTarget) Stop() {

	cmd := tui.Command(
		"Stopping pod...",
		"kubectl",
		"delete",
		"pod",
		t.serviceName)

	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr

	cmd.Run()
}
