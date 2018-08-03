package soapboxd

import (
	"database/sql"
	"net/http"

	"github.com/adhocteam/soapbox"
	"github.com/adhocteam/soapbox/proto"
)

type server struct {
	db                 *sql.DB
	httpClient         *http.Client
	config             *soapbox.Config
	configurationStore ConfigurationStore
	objectStore        ObjectStore
}

// CloudProvider is a summation type of all the interfaces that a cloud must provide
type CloudProvider interface {
	ObjectStore
	ConfigurationStore
}

// ObjectStore represents a blob store that can hold arbitrary files
type ObjectStore interface {
	UploadFile(bucket string, key string, filename string) error
}

// ConfigurationStore represents a place that can store and retrieve configurations for applications
type ConfigurationStore interface {
	GetConfigVars(appSlug string, envSlug string, version int32) ([]*proto.ConfigVar, error)
	SaveConfigVars(appSlug string, envSlug string, version int32, configVars []*proto.ConfigVar, kmsKeyARN string) error
	DeleteConfigVars(appSlug string, envSlug string, version int32) error
}

// NewServer creates an new instance of the server object
func NewServer(db *sql.DB, httpClient *http.Client, config *soapbox.Config, cloud CloudProvider) *server {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &server{
		db:                 db,
		httpClient:         httpClient,
		config:             config,
		configurationStore: cloud,
		objectStore:        cloud,
	}
}
