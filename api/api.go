package api

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/adhocteam/soapbox/api/proto"
	"golang.org/x/net/context"
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
	// TODO(paulsmith): use codegen or other helper here for
	// mapping columns and structs
	createAppSQL = `
INSERT INTO applications (name, slug, description, external_dns, github_repo_url, dockerfile_path, entrypoint_override, type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id
`
)

func (s *server) CreateApplication(ctx context.Context, req *proto.CreateApplicationRequest) (*proto.CreateApplicationResponse, error) {
	// verify access to the GitHub repo (if private, then need
	// OAuth2 token: this is not the responsibility of this
	// module, the caller should supply this server with an HTTP
	// client configured with the token)
	err := canAccessURL(s.httpClient, req.GetGithubRepoURL())
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to GitHub repo: %v", err)
	}

	// supply a default Dockerfile path ("Dockerfile")
	dockerfilePath := req.GetDockerfilePath()
	if dockerfilePath == "" {
		dockerfilePath = "Dockerfile"
	}

	args := []interface{}{
		req.GetName(),
		slugify(req.GetName()),
		req.GetDescription(),
		req.GetExternalDNS(),
		req.GetGithubRepoURL(),
		dockerfilePath,
		req.GetEntrypointOverride(),
	}

	// TODO(paulsmith): define an enum type
	// the strings must stay in sync with app_type enum in /db/schema.sql
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
		Id: int32(id),
	}

	return appResp, nil
}

type httpHead interface {
	Head(url string) (*http.Response, error)
}

func canAccessURL(client httpHead, url string) error {
	resp, err := client.Head(url)
	if err != nil {
		// TODO(paulsmith): use github.com/pkg/errors errors.Wrap instead
		return fmt.Errorf("couldn't make HTTP HEAD request to %s: %v", url, err)
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("non-success HTTP status response from %s: %d", url, resp.StatusCode)
	}
	return nil
}

var (
	slugSpaceRe      = regexp.MustCompile(`\s+`)
	slugNotAllowedRe = regexp.MustCompile(`[^a-z0-9-]`)
	slugRepeatDashRe = regexp.MustCompile(`-{2,}`)
)

func slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = slugSpaceRe.ReplaceAllString(s, "-")
	s = slugNotAllowedRe.ReplaceAllString(s, "")
	s = slugRepeatDashRe.ReplaceAllString(s, "-")
	return s
}
