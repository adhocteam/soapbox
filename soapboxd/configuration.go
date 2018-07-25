package soapboxd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/adhocteam/soapbox/proto"
	gpb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

const soapboxConfigsBucket string = "soapbox-app-configs"

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

	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "creating new session")
	}

	configObjectKey, err := s.GetS3ObjectKey(ctx, envID, config.Version)
	if err != nil {
		return nil, errors.Wrap(err, "getting s3 object key")
	}

	encryptedConfig, err := newS3Storage(sess).downloadFile(soapboxConfigsBucket, configObjectKey)
	if err != nil {
		return nil, errors.Wrap(err, "downloading config from s3")
	}

	serializedConfig, err := newKMSClient(sess).decrypt(encryptedConfig)
	if err != nil {
		return nil, errors.Wrap(err, "decrypting config")
	}

	if json.Unmarshal([]byte(serializedConfig), &config.ConfigVars) != nil {
		return nil, errors.Wrap(err, "parsing json")
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

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "committing transaction")
	}

	setPbTimestamp(config.CreatedAt, createdAt)

	err = s.UploadConfigurationsToS3(ctx, req, config)
	if err != nil {
		return nil, errors.Wrap(err, "uploading configurations to S3")
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
		return nil, errors.Wrap(err, "deleting config_vars")
	}

	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "creating new aws session")
	}

	key, err := s.GetS3ObjectKey(ctx, envID, version)
	if err != nil {
		return nil, errors.Wrap(err, "getting s3 object key")
	}

	err = newS3Storage(sess).deleteFile(soapboxConfigsBucket, key)
	if err != nil {
		return nil, errors.Wrap(err, "deleting file")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "committing transaction")
	}

	return &proto.Empty{}, nil
}

func (s *server) GetS3ObjectKey(ctx context.Context, envID int32, version int32) (string, error) {
	env, err := s.GetEnvironment(ctx, &proto.GetEnvironmentRequest{Id: envID})
	if err != nil {
		return "", errors.Wrap(err, "getting environment")
	}

	app, err := s.GetApplication(ctx, &proto.GetApplicationRequest{Id: env.GetApplicationId()})
	if err != nil {
		return "", errors.Wrap(err, "getting application")
	}

	appSlug := app.GetSlug()
	envSlug := env.GetSlug()
	return fmt.Sprintf("%s/%s/%s-configuration-v%d.json", appSlug, envSlug, appSlug, version), nil

}

func (s *server) UploadConfigurationsToS3(ctx context.Context, req *proto.CreateConfigurationRequest, config *proto.Configuration) error {
	// First, get all of the application data.
	envID := req.GetEnvironmentId()
	env, err := s.GetEnvironment(ctx, &proto.GetEnvironmentRequest{Id: envID})
	if err != nil {
		return errors.Wrap(err, "getting environment")
	}
	app, err := s.GetApplication(ctx, &proto.GetApplicationRequest{Id: env.GetApplicationId()})
	if err != nil {
		return errors.Wrap(err, "getting application")
	}

	// Encode configurations to JSON.
	serializedConfigs, err := json.Marshal(config.ConfigVars)
	if err != nil {
		return errors.Wrap(err, "serializing configurations to JSON")
	}

	// Create a new AWS session.
	sess, err := session.NewSession()
	if err != nil {
		return errors.Wrap(err, "creating new session")
	}

	// Encrypt the JSON using KMS
	// The encryption key is stored in the applications table of the database.
	encryptedConfigs, err := newKMSClient(sess).encrypt(app.GetAwsEncryptionKeyArn(), serializedConfigs)
	if err != nil {
		return errors.Wrap(err, "encrypting configurations")
	}

	tmpfile, err := ioutil.TempFile("", "config")
	if err != nil {
		return errors.Wrap(err, "creating temporary file")
	}

	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(encryptedConfigs.CiphertextBlob); err != nil {
		return errors.Wrap(err, "writing encrypted config to temporary file")
	}

	// Upload the encrypted configs JSON file to S3
	configObjectKey, err := s.GetS3ObjectKey(ctx, envID, config.GetVersion())
	if err != nil {
		return errors.Wrap(err, "getting s3 object key")
	}

	return newS3Storage(sess).uploadFile(soapboxConfigsBucket, configObjectKey, tmpfile.Name())
}
