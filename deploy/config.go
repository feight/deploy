package deploy

type Env struct {
	Config map[string]*Config
}

type Config struct {
	key          string
	Schema       string              `json:"$schema"`
	Services     map[string]*Service `required:"true"`
	UseTurboRepo bool                `description:"Use Turbo Repo to perform build."`
	GlobalEnv    []string            `description:"Global environment variables for all services."`
}

type Service struct {
	key        string
	Name       string             `required:"true" description:"Name of deployment."`
	Path       string             `required:"false" description:"Path to service. This will be the working directory."`
	Dockerfile string             `description:"Path to Dockerfile. Defaults to the working directory."`
	Prebuild   string             `description:"Pre deploy command."`
	Postdeploy string             `description:"Post deploy command."`
	Open       string             `description:"Open URL after deployment."`
	Targets    map[string]*Target `required:"true"`
}

type Target struct {
	Cloudrun          *CloudRunTarget      `description:"Use Cloud Run as target."`
	Kube              *KubernetesTarget    `description:"Use Kubernetes Engine as target."`
	Registry          *ImageRegistryTarget `description:"Do not deploy, just push to image registry."`
	CloudLoadBalancer *LoadBalancerTarget  `description:"Use Cloud Load Balancer as target."`
}

type GoogleTarget struct {
	Region      string   `required:"true" enum:"africa-south1,europe-west1"`
	ProjectId   string   `required:"true"`
	Environment []string `description:"Environment variables available at runtime."`
}

type CloudRunTarget struct {
	GoogleTarget
	UseHttp2          bool     `description:"Enable HTTP2 end-to-end. Please see https://cloud.google.com/run/docs/configuring/http2."`
	CloudSqlInstances []string `description:"Append the given values to the current Cloud SQL instances."`
	Secrets           []string `description:"List of key-value pairs to set as secrets."`
	VpcConnector      string   `description:"Set a VPC connector for this resource."`
	Cpu               string   `enum:"1,2,4,8" description:"Set a CPU limit in Kubernetes cpu units."`
	Memory            string   `enum:"512Mi,1Gi,2Gi,4Gi,8Gi" description:"Set a memory limit."`
	Concurrency       int      `description:"Set the maximum number of concurrent requests allowed per container instance. Leave concurrency unspecified or provide the special value 'default' to receive the server default value."`
	MaxInstances      int      `description:"The maximum number of container instances for this Revision to run or 'default' to remove. This setting is immutably set on each new Revision and modifying its value will deploy another Revision."`
}

type KubernetesTarget struct {
	GoogleTarget
}

type ImageRegistryTarget struct {
	GoogleTarget
}

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
