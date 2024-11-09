package deploy

import "fmt"

func (t *ImageRegistryTarget) Text() string {
	return fmt.Sprintf("[%s, Google Artifact Registry]", t.ProjectId)
}

func (t *ImageRegistryTarget) Deploy(s *Service) {
}

func (t *ImageRegistryTarget) PostDeploy(s *Service) {
}
