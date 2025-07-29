package google

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/feight/deploy/tui"
	"github.com/pkg/errors"
)

type GCETarget struct {
	GoogleTarget
	Zone         string `required:"true"`
	InstanceName string `required:"true"`
}

func (t *GCETarget) Text() string {
	return fmt.Sprintf("[%s, Google Compute Engine]", t.ProjectId)
}

func (t *GCETarget) Deploy() {

	t.Stop()

	dockerCmd := exec.Command(
		"sudo",
		"docker",
		"run", "-d",
		"--name", t.serviceName,
		"--restart=always",
		"--pull=always",
	)

	for _, v := range t.Environment {
		dockerCmd.Args = append(dockerCmd.Args, []string{"--env", v}...)
	}

	dockerCmd.Args = append(dockerCmd.Args, t.GetImageTag())

	cmd := exec.Command(
		"gcloud",
		"compute",
		"ssh",
		"--zone", t.Zone, t.InstanceName,
		"--project", t.ProjectId,
		"--command", dockerCmd.String(),
	)

	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete deploy"))
	}
}

func (t *GCETarget) PostDeploy() {

	fmt.Printf("> Tailing logs...\n\n")

	/* Wait 2s for the pod to start... */
	time.Sleep(time.Second * 2)

	t.TailLogs()
}

func (t *GCETarget) TailLogs() {

	cmd := exec.Command(
		"gcloud",
		"compute",
		"ssh",
		"--zone", t.Zone, t.InstanceName,
		"--project", t.ProjectId,
		"--command", exec.Command("sudo", "docker", "logs", "-f", t.serviceName).String(),
	)

	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}

func (t *GCETarget) Stop() {

	cmd := tui.Command(
		"Stopping pod...",
		"gcloud",
		"compute",
		"ssh",
		"--zone", t.Zone, t.InstanceName,
		"--project", t.ProjectId,
		"--command", fmt.Sprintf("%s && %s",
			exec.Command("sudo", "docker", "stop", t.serviceName).String(),
			exec.Command("sudo", "docker", "rm", t.serviceName).String()),
	)

	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr

	cmd.Run()
}
