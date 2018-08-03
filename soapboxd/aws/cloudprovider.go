package aws

import (
	"log"

	"github.com/adhocteam/soapbox/proto"
	"github.com/adhocteam/soapbox/soapboxd"
	"github.com/aws/aws-sdk-go/aws/session"
)

// This file fulfills the public interface to a cloud provider.
// The rest of this package is all private implementation details

type aws struct {
	s3     *s3storage
	kms    *kmsClient
	config *s3ConfigurationStore
}

// NewCloudProvider returns a struct that fulfills all of the cloud provider interfaces
func NewCloudProvider() soapboxd.CloudProvider {
	// create AWS session (read aws config)
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("couldn't load aws config")
	}

	s3 := newS3Storage(sess)
	kms := newKMSClient(sess)

	return &aws{
		s3:     s3, // create s3storage service. it feels a little weird to pass this to the configStore and have it on the server struct, but it's used directly during deployments for storing docker images
		kms:    kms,
		config: newS3ConfigurationStore(s3, "soapbox-app-configs", kms), // storage for application-environment level sets of configuration variables
	}
}

func (a *aws) UploadFile(bucket string, key string, filename string) error {
	return a.s3.uploadFile(bucket, key, filename)
}

func (a *aws) GetConfigVars(appSlug string, envSlug string, version int32) ([]*proto.ConfigVar, error) {
	return a.config.getConfigVars(appSlug, envSlug, version)
}

func (a *aws) SaveConfigVars(appSlug string, envSlug string, version int32, configVars []*proto.ConfigVar, kmsKeyARN string) error {
	return a.config.saveConfigVars(appSlug, envSlug, version, configVars, kmsKeyARN)
}

func (a *aws) DeleteConfigVars(appSlug string, envSlug string, version int32) error {
	return a.config.deleteConfigVars(appSlug, envSlug, version)
}
