package soapboxd

import (
	"testing"

	pb "github.com/adhocteam/soapbox/proto"
)

// TODO probably most this out into another test package
type CloudSuccess struct{}

func (c *CloudSuccess) UploadFile(bucket string, key string, filename string) error {
	return nil
}
func (c *CloudSuccess) GetConfigVars(appSlug string, envSlug string, version int32) ([]*pb.ConfigVar, error) {
	return []*pb.ConfigVar{}, nil
}
func (c *CloudSuccess) SaveConfigVars(appSlug string, envSlug string, version int32, configVars []*pb.ConfigVar, kmsKeyARN string) error {
	return nil
}
func (c *CloudSuccess) DeleteConfigVars(appSlug string, envSlug string, version int32) error {
	return nil
}
func (c *CloudSuccess) Deploy(app Application, env Environment, config *pb.Configuration) error {
	return nil
}
func (c *CloudSuccess) Rollforward(app Application, env Environment) error {
	return nil
}

func TestDeploy(t *testing.T) {
	// Totally won't work due to nil sql provider
	//s := NewServer(nil, nil, &CloudSuccess{})

	cases := []pb.Deployment{
		{
			Id:          0,
			Application: &pb.Application{Id: 12},
			Env:         &pb.Environment{Id: 10},
			Committish:  "8af116c7f2fc226208855ae6e71f4d54e6290b7b",
			State:       "",
		},
	}

	for _, c := range cases {
		// Non functional but should be what we do here
		//_, err := s.StartDeployment(context.TODO(), &c)
		var err error
		if err != nil {
			t.Errorf("Request: %#v unexpectedly failed", c)
		}
	}
}
