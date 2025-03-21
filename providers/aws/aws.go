package aws

import "fmt"

type AwsTarget struct {
	key         string
	serviceName string
	Region      string   `required:"true" enum:"africa-south1,europe-west1"`
	ProjectId   string   `required:"true"`
	Environment []string `description:"Environment variables available at build time and runtime."`
}

func (s *AwsTarget) SetKey(key string) {
	s.key = key
}

func (s *AwsTarget) Configure(serviceName string, env []string) {
	s.serviceName = serviceName
}

func (s *AwsTarget) GetRegion() string {
	return s.Region
}

func (s *AwsTarget) GetProject() string {
	return s.ProjectId
}

func (s *AwsTarget) GetEnvironment() []string {
	return s.Environment
}

func (s *AwsTarget) GetImageRegistry() string {
	return fmt.Sprintf("%s-docker.pkg.dev", s.GetRegion())
}

func (s *AwsTarget) GetImageTag() string {
	return fmt.Sprintf("%s/%s/newsteam/%s", s.GetImageRegistry(), s.GetProject(), s.serviceName)
}
