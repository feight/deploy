package deploy

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func (t *CloudRunTarget) Text() string {
	return fmt.Sprintf("[%s, Cloud Run]", t.ProjectId)
}

func (t *CloudRunTarget) Deploy(s *Service) {

	cmd := exec.Command(
		"gcloud",
		"run",
		"deploy",
		s.key,
		"--project", t.ProjectId,
		"--region", string(t.Region),
		"--platform", "managed",
		"--image", t.GetImageTag(s),
		"--allow-unauthenticated",
		"--max-instances", cmp.Or(strconv.Itoa(t.MaxInstances), "2"),
		"--concurrency", cmp.Or(strconv.Itoa(t.Concurrency), "10"),
		"--memory", cmp.Or(t.Memory, "512Mi"),
		"--cpu", cmp.Or(t.Cpu, "1"))

	cmd.Args = append(cmd.Args, []string{
		"--set-env-vars", strings.Join(env(t.Environment), ",")}...)

	if t.UseHttp2 {
		cmd.Args = append(cmd.Args, "--use-http2")
	} else {
		cmd.Args = append(cmd.Args, "--no-use-http2")
	}
	if t.VpcConnector != "" {
		cmd.Args = append(cmd.Args, "--vpc-connector", t.VpcConnector)
	}
	for _, val := range t.Secrets {
		cmd.Args = append(cmd.Args, fmt.Sprintf("--update-secrets=%s", val))
	}
	for _, val := range t.CloudSqlInstances {
		cmd.Args = append(cmd.Args, fmt.Sprintf("--add-cloudsql-instances=%s", val))
	}

	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		panic(errors.Wrap(err, "could not complete deploy"))
	}
}

func (t *CloudRunTarget) PostDeploy(s *Service) {
}
