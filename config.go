package main

import (
	"github.com/feight/deploy/providers/aws"
	"github.com/feight/deploy/providers/google"
)

type Env struct {
	Config map[string]*Config
}

type Config struct {
	key          string
	Name         string
	IsProduction bool
	Schema       string              `json:"$schema"`
	Services     map[string]*Service `required:"true"`
	UseTurboRepo bool                `description:"Use Turbo Repo to perform build."`
	GlobalEnv    []string            `description:"Global environment variables for all services."`
}

type Service struct {
	key        string
	Name       string  `required:"true" description:"Name of deployment."`
	Path       string  `required:"false" description:"Path to service. This will be the working directory."`
	Dockerfile string  `description:"Path to Dockerfile. Defaults to the working directory."`
	Prebuild   string  `description:"Pre deploy command."`
	Postdeploy string  `description:"Post deploy command."`
	Open       string  `description:"Open URL after deployment."`
	Targets    *Target `required:"true"`
}

type Target struct {
	Cloudrun          *google.CloudRunTarget         `description:"Use Cloud Run as target."`
	Kube              *google.KubernetesTarget       `description:"Use Kubernetes Engine as target."`
	GCE               *google.GCETarget              `description:"Use Google Compute Engine as target."`
	Registry          *google.ArtifactRegistryTarget `description:"Do not deploy, just push to image registry."`
	CloudLoadBalancer *google.LoadBalancerTarget     `description:"Use Cloud Load Balancer as target."`
	Lambda            *aws.LambdaTarget              `description:"Use AWS Lambda as target."`
}
