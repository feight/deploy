package deploy

import "fmt"

func (s *GoogleTarget) GetRegion() string {
	return s.Region
}

func (s *GoogleTarget) GetProject() string {
	return s.ProjectId
}

func (s *GoogleTarget) GetImageTag(service *Service) string {
	return fmt.Sprintf("%s-docker.pkg.dev/%s/newsteam/%s", s.GetRegion(), s.GetProject(), service.key)
}
