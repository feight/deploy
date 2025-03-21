package google

import "fmt"

type ArtifactRegistryTarget struct {
	GoogleTarget
}

func (t *ArtifactRegistryTarget) Text() string {
	return fmt.Sprintf("[%s, Google Artifact Registry]", t.ProjectId)
}

func (t *ArtifactRegistryTarget) Deploy() {
}

func (t *ArtifactRegistryTarget) PostDeploy() {
}
