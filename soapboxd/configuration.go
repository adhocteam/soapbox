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
	var configVars []*proto.ConfigVar
	query = `
SELECT name, value
FROM config_vars
WHERE environment_id = $1 AND version = $2
`
	rows, err := s.db.Query(query, envID, config.Version)
	if err != nil {
		return nil, errors.Wrap(err, "querying config_vars table")
	}
	for rows.Next() {
		configVar := &proto.ConfigVar{}
		dest := []interface{}{
			&configVar.Name,
			&configVar.Value,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}
		configVars = append(configVars, configVar)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "db iteration")
	}
	config.ConfigVars = configVars
	return config, nil
}

func (s *server) CreateConfiguration(ctx context.Context, req *proto.CreateConfigurationRequest) (*proto.Configuration, error) {
	envID := req.GetEnvironmentId()
	var newVersion int32
	latest, err := s.GetLatestConfiguration(ctx, &proto.GetLatestConfigurationRequest{
		EnvironmentId: envID,
	})
	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.NotFound {
				newVersion = 1
			}
		} else {
			return nil, errors.Wrap(err, "getting latest configuration")
		}
	} else {
		newVersion = latest.GetVersion() + 1
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "beginning transaction")
	}
	configVarsQuery := `
INSERT INTO config_vars (environment_id, version, name, value) 
VALUES ($1, $2, $3, $4)
`
	stmt, err := tx.Prepare(configVarsQuery)
	if err != nil {
		return nil, errors.Wrap(err, "preparing insert query")
	}

	configQuery := `
INSERT INTO configurations (environment_id, version) 
VALUES ($1, $2) 
RETURNING created_at`
	var createdAt time.Time
	if err := tx.QueryRow(configQuery, envID, newVersion).Scan(&createdAt); err != nil {
		return nil, errors.Wrap(err, "inserting into configurations table")
	}

	for _, cv := range req.ConfigVars {
		args := []interface{}{
			envID,
			newVersion,
			cv.GetName(),
			cv.GetValue(),
		}
		if _, err := stmt.Exec(args...); err != nil {
			return nil, errors.Wrap(err, "inserting into config_vars table")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "committing transaction")
	}

	resp := &proto.Configuration{
		EnvironmentId: envID,
		Version:       newVersion,
		ConfigVars:    req.ConfigVars,
		CreatedAt:     new(gpb.Timestamp),
	}
	setPbTimestamp(resp.CreatedAt, createdAt)

	return resp, nil
}

func (s *server) DeleteConfiguration(ctx context.Context, req *proto.DeleteConfigurationRequest) (*proto.Empty, error) {
	envID := req.GetEnvironmentId()
	version := req.GetVersion()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "beginning transaction")
	}

	query := `DELETE FROM config_vars WHERE environment_id = $1 AND version = $2`
	if _, err := tx.Exec(query, envID, version); err != nil {
		return nil, errors.Wrap(err, "deleting config_vars")
	}

	query = `DELETE FROM configurations WHERE environment_id = $1 AND version = $2`
	if _, err := tx.Exec(query, envID, version); err != nil {
		return nil, errors.Wrap(err, "deleting config_vars")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "committing transaction")
	}

	return &proto.Empty{}, nil
}
