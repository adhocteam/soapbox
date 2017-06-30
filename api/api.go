package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/adhocteam/soapbox/api/proto"
)

type server struct {
	db         *sql.DB
	httpClient *http.Client
}

func NewServer(db *sql.DB, httpClient *http.Client) *server {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &server{
		db:         db,
		httpClient: httpClient,
	}
}

const (
	// TODO(paulsmith): use codegen or other helper here for mapping columns and structs
	createAppSQL = `
INSERT INTO applications (name, description, external_dns, github_repo_url, dockerfile_path, entrypoint, type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id
`
)

func (s *server) CreateApplication(ctx context.Context, req *proto.CreateApplicationRequest) (*proto.CreateApplicationResponse, error) {
	// verify access the GitHub repo (private - then need OAuth2 token)
	resp, err := s.httpClient.Head(req.GetGithubRepoURL())
	if err != nil {
		// TODO(paulsmith): use github.com/pkg/errors errors.Wrap instead
		return nil, fmt.Errorf("error checking for GitHub repo: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("error checking for GitHub repo: HTTP status %d", resp.StatusCode)
	}
	// supply a default Dockerfile path ("/Dockerfile")
	dockerfilePath := req.GetDockerfilePath()
	if dockerfilePath == "" {
		dockerfilePath = "./Dockerfile"
	}

	args := []interface{}{
		req.GetName(),
		req.GetDescription(),
		req.GetExternalDNS(),
		req.GetGithubRepoURL(),
		req.GetDockerfilePath(),
		req.GetEntrypoint(),
	}

	// TODO(paulsmith): define an enum type in db and maybe in Go
	var appType string
	switch req.GetType() {
	case proto.CreateApplicationRequest_SERVER:
		appType = "server"
	case proto.CreateApplicationRequest_CRONJOB:
		appType = "cronjob"
	}
	args = append(args, appType)

	var id int

	if err := s.db.QueryRow(createAppSQL, args...).Scan(&id); err != nil {
		return nil, fmt.Errorf("couldn't insert to database new application: %v", err)
	}

	appResp := &proto.CreateApplicationResponse{
		Id: strconv.Itoa(id),
	}

	return appResp, nil
}
