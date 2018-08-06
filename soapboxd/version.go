package soapboxd

import (
	"github.com/adhocteam/soapbox/buildinfo"
	pb "github.com/adhocteam/soapbox/proto"
	"golang.org/x/net/context"
)

func (s *Server) GetVersion(ctx context.Context, req *pb.Empty) (*pb.GetVersionResponse, error) {
	return &pb.GetVersionResponse{
		Version:   buildinfo.Version,
		GitCommit: buildinfo.GitCommit,
		BuildTime: buildinfo.BuildTime,
	}, nil
}
