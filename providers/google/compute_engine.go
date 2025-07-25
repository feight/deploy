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

	// t.Stop()

	cmd := exec.Command(
		"gcloud",
		"compute",
		"ssh",
		"--zone", t.Zone, t.InstanceName,
		"--project", t.ProjectId,
		"--command", fmt.Sprintf("sudo docker run --name %s --restart=always --pull=always %s", t.serviceName, t.GetImageTag()),
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

func (t *GCETarget) PostDeploy() {

	fmt.Printf("> Waiting for logs...\n\n")

	/* Wait 10s for the pod to start... */
	// time.Sleep(time.Second * 10)

	// t.TailLogs()
}

func (t *GCETarget) TailLogs() {

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

// func (t *GCETarget) Stop() {
//
// 	cmd := tui.Command(
// 		"Stopping pod...",
// 		"kubectl",
// 		"delete",
// 		"pod",
// 		t.serviceName)
//
// 	cmd.Env = os.Environ()
// 	cmd.Stderr = os.Stderr
//
// 	cmd.Run()
// }
