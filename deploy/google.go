package deploy

import "fmt"

func (s *GoogleTarget) SetKey(key string) {
	s.key = key
}

func (s *GoogleTarget) GetRegion() string {
	return s.Region
}

func (s *GoogleTarget) GetProject() string {
	return s.ProjectId
}

func (s *GoogleTarget) GetEnvironment(*Service) []string {
	return s.Environment
}

func (s *GoogleTarget) GetImageRegistry() string {
	return fmt.Sprintf("%s-docker.pkg.dev", s.GetRegion())
}

func (s *GoogleTarget) GetImageTag(service *Service) string {
	return fmt.Sprintf("%s/%s/newsteam/%s", s.GetImageRegistry(), s.GetProject(), service.key)
}
