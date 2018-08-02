package soapboxd

import (
	"database/sql"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/adhocteam/soapbox/proto"
	gpb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func (s *server) ListConfigurations(ctx context.Context, req *proto.ListConfigurationRequest) (*proto.ListConfigurationResponse, error) {
	// TODO(paulsmith): maybe embed actual config vars in response payload
	query := `SELECT version, created_at FROM configurations WHERE environment_id = $1`
	var configs []*proto.Configuration
	envID := req.GetEnvironmentId()
	rows, err := s.db.Query(query, envID)
	if err != nil {
		return nil, errors.Wrap(err, "querying configurations table")
	}
	for rows.Next() {
		c := &proto.Configuration{
			EnvironmentId: envID,
		}
		var createdAt time.Time
		if err := rows.Scan(&c.Version, &createdAt); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}
		setPbTimestamp(c.CreatedAt, createdAt)
		configs = append(configs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "db iteration")
	}
	resp := &proto.ListConfigurationResponse{
		Configs: configs,
	}
	return resp, nil
}

func (s *server) GetLatestConfiguration(ctx context.Context, req *proto.GetLatestConfigurationRequest) (*proto.Configuration, error) {
	// TODO(paulsmith): FIXME return error or message with error
	// semantics if we get nothing back from the db, instead of
	// returning a zero-value configuration (in the case where an
	// environment doesn't have any configurations)
	query := `
SELECT version, created_at
FROM configurations
WHERE environment_id = $1
ORDER BY version DESC
LIMIT 1
`
	config := &proto.Configuration{
		CreatedAt: new(gpb.Timestamp),
	}
	envID := req.GetEnvironmentId()
	var createdAt time.Time
	if err := s.db.QueryRow(query, envID).Scan(&config.Version, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "no configuration found for environment")
		} else {
			return nil, errors.Wrap(err, "querying configurations table")
		}
	}
	setPbTimestamp(config.CreatedAt, createdAt)

	env, err := s.GetEnvironment(ctx, &proto.GetEnvironmentRequest{Id: envID})
	if err != nil {
		return nil, errors.Wrap(err, "getting environment")
	}

	app, err := s.GetApplication(ctx, &proto.GetApplicationRequest{Id: env.GetApplicationId()})
	if err != nil {
		return nil, errors.Wrap(err, "getting application")
	}

	appSlug := app.GetSlug()
	envSlug := env.GetSlug()

	config.ConfigVars, err = s.configurationStore.GetConfigVars(appSlug, envSlug, config.Version)
	if err != nil {
		return nil, errors.Wrap(err, "getting config variables")
	}

	return config, nil
}

func (s *server) CreateConfiguration(ctx context.Context, req *proto.CreateConfigurationRequest) (*proto.Configuration, error) {
	envID := req.GetEnvironmentId()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "beginning transaction")
	}

	config := &proto.Configuration{
		EnvironmentId: envID,
		ConfigVars:    req.ConfigVars,
		CreatedAt:     new(gpb.Timestamp),
	}

	configQuery := `
INSERT INTO configurations (environment_id)
VALUES ($1)
RETURNING created_at, version`

	var createdAt time.Time
	if err := tx.QueryRow(configQuery, envID).Scan(&createdAt, &config.Version); err != nil {
		return nil, errors.Wrap(err, "inserting into configurations table")
	}

	setPbTimestamp(config.CreatedAt, createdAt)

	env, err := s.GetEnvironment(ctx, &proto.GetEnvironmentRequest{Id: envID})
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "getting environment")
	}

	app, err := s.GetApplication(ctx, &proto.GetApplicationRequest{Id: env.GetApplicationId()})
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "getting application")
	}

	kmsKeyARN := app.GetAwsEncryptionKeyArn()

	appSlug := app.GetSlug()
	envSlug := env.GetSlug()

	err = s.configurationStore.SaveConfigVars(appSlug, envSlug, config.Version, config.ConfigVars, kmsKeyARN)
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "saving config variables")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "committing transaction")
	}

	return config, nil
}

func (s *server) DeleteConfiguration(ctx context.Context, req *proto.DeleteConfigurationRequest) (*proto.Empty, error) {
	envID := req.GetEnvironmentId()
	version := req.GetVersion()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "beginning transaction")
	}

	query := `DELETE FROM configurations WHERE environment_id = $1 AND version = $2`
	if _, err := tx.Exec(query, envID, version); err != nil {
		return nil, errors.Wrap(err, "deleting configuration from local database")
	}

	env, err := s.GetEnvironment(ctx, &proto.GetEnvironmentRequest{Id: envID})
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "getting environment")
	}

	app, err := s.GetApplication(ctx, &proto.GetApplicationRequest{Id: env.GetApplicationId()})
	if err != nil {
		tx.Rollback()
		return nil, errors.Wrap(err, "getting application")
	}

	appSlug := app.GetSlug()
	envSlug := env.GetSlug()

	err = s.configurationStore.DeleteConfigVars(appSlug, envSlug, version)
	if err != nil {
		return nil, errors.Wrap(err, "deleting config variables")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "committing transaction")
	}

	return &proto.Empty{}, nil
}
