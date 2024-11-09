package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

func (t *LoadBalancerTarget) Text() string {
	return fmt.Sprintf("[%s, Google Cloud Load Balancer]", t.ProjectId)
}

func (t *LoadBalancerTarget) Deploy(s *Service) {

	b, _ := json.Marshal(t.LoadBalancerTargetRules)

	cmd := exec.Command(
		"gcloud",
		"compute",
		"url-maps",
		"import",
		"-q",
		t.Name,
		"--project", t.ProjectId)

	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr

	stdin, _ := cmd.StdinPipe()

	err := cmd.Start()

	if err != nil {
		panic(errors.Wrap(err, "could not complete load balancer deploy"))
	}

	stdin.Write(b)
	stdin.Close()

	cmd.Wait()
}

func (t *LoadBalancerTarget) PostDeploy(s *Service) {
}
