package aws

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/adhocteam/soapbox/proto"

	"github.com/pkg/errors"
)

type s3ConfigurationStore struct {
	s3     *s3storage
	bucket string
	kms    *kmsClient
}

func newS3ConfigurationStore(s3 *s3storage, bucket string, kms *kmsClient) *s3ConfigurationStore {
	return &s3ConfigurationStore{
		s3:     s3,
		bucket: bucket,
		kms:    kms,
	}
}

func (s *s3ConfigurationStore) getConfigVars(appSlug string, envSlug string, version int32) ([]*proto.ConfigVar, error) {
	encryptedConfig, err := s.s3.downloadFile(s.bucket, getS3ObjectKey(appSlug, envSlug, version))
	if err != nil {
		return nil, errors.Wrap(err, "retrieving config from s3")
	}

	serializedConfig, err := s.kms.decrypt(encryptedConfig)
	if err != nil {
		return nil, errors.Wrap(err, "decrypting config")
	}

	var configVars []*proto.ConfigVar

	if json.Unmarshal([]byte(serializedConfig), &configVars) != nil {
		return nil, errors.Wrap(err, "parsing json")
	}

	return configVars, nil

}

func (s *s3ConfigurationStore) saveConfigVars(appSlug string, envSlug string, version int32, configVars []*proto.ConfigVar, kmsKeyARN string) error {
	serializedConfigs, err := json.Marshal(configVars)
	if err != nil {
		return errors.Wrap(err, "serializing configurations to JSON")
	}

	encryptedConfigs, err := s.kms.encrypt(kmsKeyARN, serializedConfigs)
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

	return s.s3.uploadFile(s.bucket, getS3ObjectKey(appSlug, envSlug, version), tmpfile.Name())
}

func (s *s3ConfigurationStore) deleteConfigVars(appSlug string, envSlug string, version int32) error {
	return s.s3.deleteFile(s.bucket, getS3ObjectKey(appSlug, envSlug, version))
}

func getS3ObjectKey(appSlug string, envSlug string, version int32) string {
	return fmt.Sprintf("%s/%s/%s-configuration-v%d.json", appSlug, envSlug, appSlug, version)
}
