package deploy

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

func (t *KubernetesTarget) Deploy(s *Service) {

	t.setCluster()
	t.Stop(s)

	cmd := exec.Command(
		"kubectl",
		"run",
		s.key,
		"--image", t.GetImageTag(s),
		"--image-pull-policy", "Always",
	)

	for _, v := range env(t.Environment) {
		cmd.Args = append(cmd.Args, []string{"--env", v}...)
	}

	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete deploy"))
	}
}

func (t *KubernetesTarget) PostDeploy(s *Service) {

	fmt.Printf("> Waiting for logs...\n\n")

	/* Wait 10s for the pod to start... */
	time.Sleep(time.Second * 10)

	t.TailLogs(s)
}

func (t *KubernetesTarget) TailLogs(s *Service) {

	cmd := exec.Command(
		"kubectl",
		"logs", "-f",
		s.key,
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

		t.TailLogs(s)
	}
}

func (t *KubernetesTarget) Stop(s *Service) {

	cmd := tui.Command(
		"Stopping pod...",
		"kubectl",
		"delete",
		"pod",
		s.key)

	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr

	cmd.Run()
}
