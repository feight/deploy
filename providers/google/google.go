package google

import "fmt"

type GoogleTarget struct {
	key         string
	serviceName string
	Region      string   `required:"true" enum:"africa-south1,europe-west1"`
	ProjectId   string   `required:"true"`
	Environment []string `description:"Environment variables available at build time and runtime."`
}

func (s *GoogleTarget) SetKey(key string) {
	s.key = key
}

func (s *GoogleTarget) SetServiceName(name string) {
	s.serviceName = name
}

func (s *GoogleTarget) GetRegion() string {
	return s.Region
}

func (s *GoogleTarget) GetProject() string {
	return s.ProjectId
}

func (s *GoogleTarget) GetEnvironment() []string {
	return s.Environment
}

func (s *GoogleTarget) GetImageRegistry() string {
	return fmt.Sprintf("%s-docker.pkg.dev", s.GetRegion())
}

func (s *GoogleTarget) GetImageTag() string {
	return fmt.Sprintf("%s/%s/newsteam/%s", s.GetImageRegistry(), s.GetProject(), s.serviceName)
}
