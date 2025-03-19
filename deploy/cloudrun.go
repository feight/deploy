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
		"--vpc-egress", "private-ranges-only",
		"--max-instances", t.getMaxInstances(),
		"--concurrency", t.getConcurrency(),
		"--memory", cmp.Or(t.Memory, "512Mi"),
		"--cpu", cmp.Or(t.Cpu, "1"))

	cmd.Args = append(cmd.Args, []string{
		"--set-env-vars", strings.Join(env(t.Environment), ",")}...)

	if t.UseHttp2 {
		cmd.Args = append(cmd.Args, "--use-http2")
	} else {
		cmd.Args = append(cmd.Args, "--no-use-http2")
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

func (t *CloudRunTarget) getConcurrency() string {

	if t.Concurrency == 0 {
		return "100"
	}
	return strconv.Itoa(t.Concurrency)
}

func (t *CloudRunTarget) getMaxInstances() string {

	if t.MaxInstances == 0 {
		return "2"
	}
	return strconv.Itoa(t.MaxInstances)
}

func (t *CloudRunTarget) PostDeploy(s *Service) {
}
