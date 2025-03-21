package google

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type LoadBalancerTarget struct {
	GoogleTarget
	LoadBalancerTargetRules
	Name string `required:"true"`
}

type LoadBalancerTargetRules struct {
	DefaultService string `json:"defaultService"`

	HostRules []struct {
		Hosts       []string `json:"hosts"`
		PathMatcher string   `json:"pathMatcher"`
	} `json:"hostRules"`

	PathMatchers []struct {
		DefaultService string `json:"defaultService"`
		Name           string `json:"name"`
	} `json:"pathMatchers"`
}

func (t *LoadBalancerTarget) Text() string {
	return fmt.Sprintf("[%s, Google Cloud Load Balancer]", t.ProjectId)
}

func (t *LoadBalancerTarget) Deploy() {

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

func (t *LoadBalancerTarget) PostDeploy() {}
