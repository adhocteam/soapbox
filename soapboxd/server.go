package soapboxd

import (
	"database/sql"
	"net/http"

	"github.com/adhocteam/soapbox"
	pb "github.com/adhocteam/soapbox/proto"
)

type server struct {
	db                 *sql.DB
	httpClient         *http.Client
	configurationStore ConfigurationStore
	objectStore        ObjectStore
	deployer           Deployer
}

// CloudProvider is a summation type of all the interfaces that a cloud must provide
type CloudProvider interface {
	NewCloudProvider(config soapbox.Config) CloudProvider
	ObjectStore
	ConfigurationStore
	Deployer
}

// ObjectStore represents a blob store that can hold arbitrary files
type ObjectStore interface {
	UploadFile(bucket string, key string, filename string) error
}

// ConfigurationStore represents a place that can store and retrieve configurations for applications
type ConfigurationStore interface {
	GetConfigVars(appSlug string, envSlug string, version int32) ([]*pb.ConfigVar, error)
	SaveConfigVars(appSlug string, envSlug string, version int32, configVars []*pb.ConfigVar, kmsKeyARN string) error
	DeleteConfigVars(appSlug string, envSlug string, version int32) error
}

// Deployer represents something that can blue/green deploy an image to a cloud provider
type Deployer interface {
	Deploy(app Application, env Environment, config *pb.Configuration) error // Deploy is defined by the trio of application, environment, and configuration
	Rollforward() error                                                      // Finalize a successful deployment
}

// NewServer creates an new instance of the server object
func NewServer(db *sql.DB, httpClient *http.Client, cloud CloudProvider) *server {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &server{
		db:                 db,
		httpClient:         httpClient,
		configurationStore: cloud,
		objectStore:        cloud,
		deployer:           cloud,
	}
}
