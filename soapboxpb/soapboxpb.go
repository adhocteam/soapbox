package soapboxpb

import (
	"fmt"
)

//go:generate protoc --proto_path=. --go_out=plugins=grpc:. soapbox.proto application.proto deployment.proto environment.proto

// helper methods

// DeploymentStateToString fetches enum string given int
func DeploymentStateToString(s DeploymentState) string {
	return DeploymentState_name[int32(s)]
}

// StringToDeploymentState fetches DeploymentState int given enum string
func StringToDeploymentState(s string) (DeploymentState, error) {
	state, exists := DeploymentState_value[s]

	if !exists {
		return -1, fmt.Errorf("Unknown deployment state: %v", s)
	}

	return DeploymentState(state), nil
}
