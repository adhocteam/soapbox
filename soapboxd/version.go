package soapboxd

import (
	pb "github.com/adhocteam/soapbox/proto"
	"github.com/adhocteam/soapbox/version"
	"golang.org/x/net/context"
)

func (s *server) GetVersion(ctx context.Context, req *pb.Empty) (*pb.GetVersionResponse, error) {
	return &pb.GetVersionResponse{
		Version:   version.Version,
		GitCommit: version.GitCommit,
		BuildTime: version.BuildTime,
	}, nil
}
