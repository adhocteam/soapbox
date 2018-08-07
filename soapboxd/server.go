package soapboxd

import (
	"database/sql"
	"net/http"

	pb "github.com/adhocteam/soapbox/proto"
)

// Server is the basic soapbox server containing all the initialized items needed to perform its functions
type Server struct {
	db                 *sql.DB
	httpClient         *http.Client
	configurationStore ConfigurationStore
	objectStore        ObjectStore
	deployer           Deployer
}

// CloudProvider is a summation type of all the interfaces that a cloud must provide
type CloudProvider interface {
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
	Rollforward(app Application, env Environment) error                      // Finalize a successful deployment
}

// NewServer creates an new instance of the server object
func NewServer(db *sql.DB, httpClient *http.Client, cloud CloudProvider) *Server {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Server{
		db:                 db,
		httpClient:         httpClient,
		configurationStore: cloud,
		objectStore:        cloud,
		deployer:           cloud,
	}
}
