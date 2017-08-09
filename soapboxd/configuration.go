package soapboxd

import (
	"github.com/adhocteam/soapbox/proto"
	"golang.org/x/net/context"
)

func (s *server) ListConfigurations(ctx context.Context, req *proto.ListConfigurationRequest) (*proto.ListConfigurationResponse, error) {
	return nil, nil
}

func (s *server) GetLatestConfiguration(ctx context.Context, req *proto.GetLatestConfigurationRequest) (*proto.Configuration, error) {
	return nil, nil
}

func (s *server) CreateConfiguration(ctx context.Context, req *proto.CreateConfigurationRequest) (*proto.Configuration, error) {
	return nil, nil
}

func (s *server) DeleteConfiguration(ctx context.Context, req *proto.DeleteConfigurationRequest) (*proto.Empty, error) {
	return nil, nil

}
