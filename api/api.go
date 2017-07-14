package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/adhocteam/soapbox/models"
	pb "github.com/adhocteam/soapbox/soapboxpb"
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

func newNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func (s *server) CreateApplication(ctx context.Context, app *pb.Application) (*pb.Application, error) {
	// verify access to the GitHub repo (if private, then need
	// OAuth2 token: this is not the responsibility of this
	// module, the caller should supply this server with an HTTP
	// client configured with the token)
	err := canAccessURL(s.httpClient, app.GetGithubRepoUrl())
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to GitHub repo: %v", err)
	}

	// supply a default Dockerfile path ("Dockerfile")
	dockerfilePath := app.GetDockerfilePath()
	if dockerfilePath == "" {
		dockerfilePath = "Dockerfile"
	}

	model := &models.Application{
		ID:                 int(app.Id),
		Name:               app.Name,
		Slug:               slugify(app.Name),
		Description:        newNullString(app.Description),
		ExternalDNS:        newNullString(app.ExternalDns),
		GithubRepoURL:      newNullString(app.GithubRepoUrl),
		DockerfilePath:     newNullString(app.DockerfilePath),
		EntrypointOverride: newNullString(app.EntrypointOverride),
		Type:               appTypePbToModel(app.Type),
		InternalDNS:        newNullString(app.InternalDns),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := model.Insert(s.db); err != nil {
		return nil, fmt.Errorf("inserting into db: %v", err)
	}

	app.Id = int32(model.ID)

	return app, nil
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

const (
	listAppsSQL = `SELECT id, name, description, created_at FROM applications ORDER BY created_at ASC`
)

func (s *server) ListApplications(ctx context.Context, _ *pb.Empty) (*pb.ListApplicationResponse, error) {
	rows, err := s.db.Query(listAppsSQL)
	if err != nil {
		return nil, fmt.Errorf("querying db for apps list: %v", err)
	}

	var apps []*pb.Application

	for rows.Next() {
		var a pb.Application
		dest := []interface{}{&a.Id, &a.Name, &a.Description, &a.CreatedAt}
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("scanning db row: %v", err)
		}
		apps = append(apps, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over db rows: %v", err)
	}

	resp := &pb.ListApplicationResponse{
		Applications: apps,
	}

	return resp, nil
}

func appTypeModelToPb(at models.AppType) pb.ApplicationType {
	switch at {
	case models.AppTypeServer:
		return pb.ApplicationType_SERVER
	case models.AppTypeCronjob:
		return pb.ApplicationType_CRONJOB
	}
	panic("shouldn't reach here")
}

func appTypePbToModel(at pb.ApplicationType) models.AppType {
	switch at {
	case pb.ApplicationType_SERVER:
		return models.AppTypeServer
	case pb.ApplicationType_CRONJOB:
		return models.AppTypeCronjob
	}
	panic("shouldn't reach here")
}

func (s *server) GetApplication(ctx context.Context, req *pb.GetApplicationRequest) (*pb.Application, error) {
	model, err := models.ApplicationByID(s.db, int(req.Id))
	if err != nil {
		return nil, fmt.Errorf("getting application by ID from db: %v", err)
	}

	app := &pb.Application{
		Id:   int32(model.ID),
		Name: model.Name,
		Slug: model.Slug,
		Type: appTypeModelToPb(model.Type),
	}

	if model.Description.Valid {
		app.Description = model.Description.String
	}
	if model.InternalDNS.Valid {
		app.InternalDns = model.InternalDNS.String
	}
	if model.ExternalDNS.Valid {
		app.ExternalDns = model.ExternalDNS.String
	}
	if model.GithubRepoURL.Valid {
		app.GithubRepoUrl = model.GithubRepoURL.String
	}
	if model.DockerfilePath.Valid {
		app.DockerfilePath = model.DockerfilePath.String
	}
	if model.EntrypointOverride.Valid {
		app.EntrypointOverride = model.EntrypointOverride.String
	}
	// TODO(paulsmith): have a global timestamp format across the
	// Go and Rails apps
	app.CreatedAt = model.CreatedAt.Format(timestampFormat)

	return app, nil
}

const timestampFormat = "2006-01-02T15:04:05"

func (s *server) ListEnvironments(ctx context.Context, req *pb.ListEnvironmentRequest) (*pb.ListEnvironmentResponse, error) {
	listSQL := "SELECT id, application_id, name, slug, vars, created_at FROM environments WHERE application_id = $1 ORDER BY id"
	rows, err := s.db.Query(listSQL, req.GetApplicationId())
	if err != nil {
		return nil, fmt.Errorf("querying db for environments: %v", err)
	}
	var envs []*pb.Environment
	for rows.Next() {
		var env pb.Environment
		var vars []byte
		dest := []interface{}{
			&env.Id,
			&env.ApplicationId,
			&env.Name,
			&env.Slug,
			&vars,
			&env.CreatedAt,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("scanning db row: %v", err)
		}
		if err := json.Unmarshal(vars, &env.Vars); err != nil {
			return nil, fmt.Errorf("unmarshalling env vars JSON: %v", err)
		}
		envs = append(envs, &env)
		log.Printf("env: %+v", env)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over db result set: %v", err)
	}
	res := &pb.ListEnvironmentResponse{Environments: envs}
	return res, nil
}

func (s *server) GetEnvironment(ctx context.Context, req *pb.GetEnvironmentRequest) (*pb.Environment, error) {
	getSQL := "SELECT id, application_id, name, slug, vars, created_at FROM environments WHERE id = $1"
	var env pb.Environment
	var vars []byte
	dest := []interface{}{
		&env.Id,
		&env.ApplicationId,
		&env.Name,
		&env.Slug,
		&vars,
		&env.CreatedAt,
	}
	if err := s.db.QueryRow(getSQL, req.GetId()).Scan(dest...); err != nil {
		return nil, fmt.Errorf("scanning db row: %v", err)
	}
	if err := json.Unmarshal(vars, &env.Vars); err != nil {
		return nil, fmt.Errorf("unmarshalling env vars JSON: %v", err)
	}
	return &env, nil
}

func (s *server) CreateEnvironment(ctx context.Context, req *pb.Environment) (*pb.Environment, error) {
	// TODO(paulsmith): can we even do this in XO??
	insertSQL := "INSERT INTO environments (application_id, name, slug, vars) VALUES ($1, $2, $3, $4) RETURNING id, created_at"

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req.Vars); err != nil {
		return nil, fmt.Errorf("encoding env vars as JSON: %v", err)
	}

	args := []interface{}{
		req.GetApplicationId(),
		req.GetName(),
		slugify(req.GetName()),
		buf.String(),
	}

	var id int

	if err := s.db.QueryRow(insertSQL, args...).Scan(&id, &req.CreatedAt); err != nil {
		return nil, fmt.Errorf("inserting in to db: %v", err)
	}

	req.Id = int32(id)

	return req, nil
}

func (s *server) DestroyEnvironment(ctx context.Context, req *pb.DestroyEnvironmentRequest) (*pb.Empty, error) {
	deleteSQL := "DELETE FROM environments WHERE id = $1"
	if _, err := s.db.Exec(deleteSQL, req.GetId()); err != nil {
		return nil, fmt.Errorf("deleting row from db: %v", err)
	}
	return &pb.Empty{}, nil
}

func (s *server) CopyEnvironment(context.Context, *pb.CopyEnvironmentRequest) (*pb.Environment, error) {
	return nil, nil
}
