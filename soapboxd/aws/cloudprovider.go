package aws

import (
	"log"

	"github.com/adhocteam/soapbox"

	pb "github.com/adhocteam/soapbox/proto"
	"github.com/adhocteam/soapbox/soapboxd"
	"github.com/aws/aws-sdk-go/aws/session"
)

// This file fulfills the public interface to a cloud provider.
// The rest of this package is all private implementation details

type aws struct {
	s3        *s3storage
	kms       *kmsClient
	appConfig *s3ConfigurationStore
	sess      *session.Session
	config    soapbox.Config
}

// NewCloudProvider returns a struct that fulfills all of the cloud provider interfaces
func NewCloudProvider(config soapbox.Config) soapboxd.CloudProvider {
	// create AWS session (read aws config)
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("couldn't load aws config")
	}

	s3 := newS3Storage(sess)
	kms := newKMSClient(sess)

	return &aws{
		s3:        s3, // create s3storage service. it feels a little weird to pass this to the configStore and have it on the server struct, but it's used directly during deployments for storing docker images
		kms:       kms,
		appConfig: newS3ConfigurationStore(s3, "soapbox-app-configs", kms), // storage for application-environment level sets of configuration variables
		sess:      sess,                                                    // Cache the config per AWS recommendation to avoid repeat startup costs
		config:    config,
	}
}

func (a *aws) UploadFile(bucket string, key string, filename string) error {
	return a.s3.uploadFile(bucket, key, filename)
}

func (a *aws) GetConfigVars(appSlug string, envSlug string, version int32) ([]*pb.ConfigVar, error) {
	return a.appConfig.getConfigVars(appSlug, envSlug, version)
}

func (a *aws) SaveConfigVars(appSlug string, envSlug string, version int32, configVars []*pb.ConfigVar, kmsKeyARN string) error {
	return a.appConfig.saveConfigVars(appSlug, envSlug, version, configVars, kmsKeyARN)
}

func (a *aws) DeleteConfigVars(appSlug string, envSlug string, version int32) error {
	return a.appConfig.deleteConfigVars(appSlug, envSlug, version)
}

func (a *aws) Deploy(app soapboxd.Application, env soapboxd.Environment, config *pb.Configuration) error {
	return a.deploy(app, env, config)
}

func (a *aws) Rollforward(app soapboxd.Application, env soapboxd.Environment) error {
	return a.rollforward(app, env)
}
