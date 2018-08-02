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

type ObjectStore interface {
	UploadFile(bucket string, key string, filename string) error
}

type ConfigurationStore interface {
	GetConfigVars(appSlug string, envSlug string, version int32) ([]*proto.ConfigVar, error)
	SaveConfigVars(appSlug string, envSlug string, version int32, configVars []*proto.ConfigVar, kmsKeyARN string) error
	DeleteConfigVars(appSlug string, envSlug string, version int32) error
}

func NewServer(db *sql.DB, httpClient *http.Client, config *soapbox.Config, configurationStore ConfigurationStore, objectStore ObjectStore) *server {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &server{
		db:                 db,
		httpClient:         httpClient,
		config:             config,
		configurationStore: configurationStore,
		objectStore:        objectStore,
	}
}
