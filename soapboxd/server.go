package soapboxd

import (
	"database/sql"
	"net/http"

	"github.com/adhocteam/soapbox"
)

type server struct {
	db         *sql.DB
	httpClient *http.Client
	config     *soapbox.Config
}

func NewServer(db *sql.DB, httpClient *http.Client, config *soapbox.Config) *server {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &server{
		db:         db,
		httpClient: httpClient,
		config:     config,
	}
}
