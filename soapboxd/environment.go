package soapboxd

import (
	"time"

	pb "github.com/adhocteam/soapbox/proto"
	gpb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"

	"golang.org/x/net/context"
)

func (s *Server) ListEnvironments(ctx context.Context, req *pb.ListEnvironmentRequest) (*pb.ListEnvironmentResponse, error) {
	listSQL := "SELECT id, application_id, name, slug, created_at FROM environments WHERE application_id = $1 ORDER BY id"
	rows, err := s.db.Query(listSQL, req.GetApplicationId())
	if err != nil {
		return nil, errors.Wrap(err, "querying db for environments")
	}
	var envs []*pb.Environment
	for rows.Next() {
		env := &pb.Environment{
			CreatedAt: new(gpb.Timestamp),
		}
		var createdAt time.Time
		dest := []interface{}{
			&env.Id,
			&env.ApplicationId,
			&env.Name,
			&env.Slug,
			&createdAt,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}
		setPbTimestamp(env.CreatedAt, createdAt)
		envs = append(envs, env)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterating over db result set")
	}
	res := &pb.ListEnvironmentResponse{Environments: envs}
	return res, nil
}

func (s *Server) GetEnvironment(ctx context.Context, req *pb.GetEnvironmentRequest) (*pb.Environment, error) {
	getSQL := "SELECT id, application_id, name, slug, created_at FROM environments WHERE id = $1"
	env := &pb.Environment{
		CreatedAt: new(gpb.Timestamp),
	}
	var createdAt time.Time
	dest := []interface{}{
		&env.Id,
		&env.ApplicationId,
		&env.Name,
		&env.Slug,
		&createdAt,
	}
	if err := s.db.QueryRow(getSQL, req.GetId()).Scan(dest...); err != nil {
		return nil, errors.Wrap(err, "scanning db row")
	}
	setPbTimestamp(env.CreatedAt, createdAt)
	return env, nil
}

func (s *Server) CreateEnvironment(ctx context.Context, req *pb.Environment) (*pb.Environment, error) {
	query := "INSERT INTO environments (application_id, name, slug) VALUES ($1, $2, $3) RETURNING id, created_at"

	args := []interface{}{
		req.GetApplicationId(),
		req.GetName(),
		slugify(req.GetName()),
	}

	var id int
	var createdAt time.Time

	if err := s.db.QueryRow(query, args...).Scan(&id, &createdAt); err != nil {
		return nil, errors.Wrap(err, "inserting in to db")
	}

	req.Id = int32(id)
	req.CreatedAt = new(gpb.Timestamp)
	setPbTimestamp(req.CreatedAt, createdAt)

	// create environment configuration
	configReq := pb.CreateConfigurationRequest{EnvironmentId: req.Id}
	_, err := s.CreateConfiguration(ctx, &configReq)
	if err != nil {
		return nil, err
	}

	if err := s.AddCreateEnvironmentActivity(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

func (s *Server) DestroyEnvironment(ctx context.Context, req *pb.DestroyEnvironmentRequest) (*pb.Empty, error) {
	deleteSQL := "DELETE FROM environments WHERE id = $1"
	if _, err := s.db.Exec(deleteSQL, req.GetId()); err != nil {
		return nil, errors.Wrap(err, "deleting row from db")
	}
	activity := pb.Activity{
		Type:          pb.ActivityType_ENVIRONMENT_DESTROYED,
		EnvironmentId: req.GetId(),
	}

	if _, err := s.AddActivity(ctx, &activity); err != nil {
		return &pb.Empty{}, err
	}
	return &pb.Empty{}, nil
}
