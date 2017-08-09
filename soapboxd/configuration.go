package soapboxd

import (
	"time"

	"github.com/adhocteam/soapbox/proto"
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
	query := `
SELECT c.version, c.created_at, cv.name, cv.value
FROM configurations c INNER JOIN config_vars cv ON c.environment_id = cv.environment_ID AND c.version = cv.version
WHERE c.environment_id = $1
ORDER BY c.version DESC
LIMIT 1
`
	config := &proto.Configuration{}
	configVars := config.ConfigVars
	envID := req.GetEnvironmentId()
	rows, err := s.db.Query(query, envID)
	if err != nil {
		return nil, errors.Wrap(err, "querying configurations / config_vars tables")
	}
	for rows.Next() {
		if config.EnvironmentId == 0 {
			config.EnvironmentId = envID
		}
		var configVar *proto.ConfigVar
		var createdAt time.Time
		var version int32
		dest := []interface{}{
			&version,
			&createdAt,
			&configVar.Name,
			&configVar.Value,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}
		if config.Version == 0 {
			config.Version = version
		}
		setPbTimestamp(config.CreatedAt, createdAt)
		configVars = append(configVars, configVar)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "db iteration")
	}
	return config, nil
}

func (s *server) CreateConfiguration(ctx context.Context, req *proto.CreateConfigurationRequest) (*proto.Configuration, error) {
	envID := req.GetEnvironmentId()
	latest, err := s.GetLatestConfiguration(ctx, &proto.GetLatestConfigurationRequest{
		EnvironmentId: envID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting latest configuration")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "beginning transaction")
	}
	configVarsQuery := `
INSERT INTO config_vars (environment_id, version, name, value) 
VALUES ($1, $2, $3, $4
`
	stmt, err := tx.Prepare(configVarsQuery)
	if err != nil {
		return nil, errors.Wrap(err, "preparing insert query")
	}

	newVersion := latest.GetVersion() + 1

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
