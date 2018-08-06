package soapboxd

import (
	"testing"

	pb "github.com/adhocteam/soapbox/proto"
	"github.com/adhocteam/soapbox/soapboxd"
)

// TODO probably most this out into another test package
type CloudSuccess struct{}

func (c *CloudSuccess) UploadFile(bucket string, key string, filename string) error {
	return nil
}
func (c *CloudSuccess) GetConfigVars(appSlug string, envSlug string, version int32) ([]*pb.ConfigVar, error) {
	return []*proto.ConfigVar{}, nil
}
func (c *CloudSuccess) SaveConfigVars(appSlug string, envSlug string, version int32, configVars []*pb.ConfigVar, kmsKeyARN string) error {
	return nil
}
func (c *CloudSuccess) DeleteConfigVars(appSlug string, envSlug string, version int32) error {
	return nil
}
func (c *CloudSuccess) Deploy(app soapboxd.Application, env soapboxd.Environment, config *pb.Configuration) error {
	return nil
}
func (c *CloudSuccess) Rollforward(app soapboxd.Application, env soapboxd.Environment) error {
	return nil
}

func TestDeploy(t *testing.T) {
	// Totally won't work due to nil sql provider
	s := NewServer(nil, nil, CloudSuccess{})

	cases := []struct {
		req *pb.Deployment
	}{
		{
			Id:          0,
			Application: &Application{},
			Env:         &Environment{},
			Committish:  "8af116c7f2fc226208855ae6e71f4d54e6290b7b",
			State:       "",
		},
	}

	for _, c := range cases {
		_, err := s.StartDeployment(ctx.TODO(), nil)
		if err != nil {
			t.Errorf("Request: %#v unexpectedly failed", c)
		}
	}
}
