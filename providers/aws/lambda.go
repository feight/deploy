package aws

import (
	"fmt"
)

type LambdaTarget struct {
	AwsTarget
}

func (t *LambdaTarget) Text() string {
	return fmt.Sprintf("[%s, AWS Lambda Serverless Function]", t.ProjectId)
}

func (t *LambdaTarget) Deploy() {
	panic("unimplemented")
}

func (t *LambdaTarget) PostDeploy() {
}
