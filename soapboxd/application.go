package soapboxd

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/adhocteam/soapbox/models"
	pb "github.com/adhocteam/soapbox/proto"
	gpb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func (s *server) CreateApplication(ctx context.Context, app *pb.Application) (*pb.Application, error) {
	// verify access to the GitHub repo (if private, then need
	// OAuth2 token: this is not the responsibility of this
	// module, the caller should supply this server with an HTTP
	// client configured with the token)
	err := canAccessURL(s.httpClient, app.GetGithubRepoUrl())
	if err != nil {
		return nil, errors.Wrap(err, "couldn't connect to Github repo")
	}

	// supply a default Dockerfile path ("Dockerfile")
	dockerfilePath := app.GetDockerfilePath()
	if dockerfilePath == "" {
		dockerfilePath = "Dockerfile"
	}

	app.Slug = slugify(app.GetName())

	model := &models.Application{
		ID:                 int(app.Id),
		UserID:             int(app.UserId),
		Name:               app.GetName(),
		Slug:               app.GetSlug(),
		Description:        newNullString(app.Description),
		ExternalDNS:        newNullString(app.ExternalDns),
		GithubRepoURL:      newNullString(app.GithubRepoUrl),
		DockerfilePath:     newNullString(app.DockerfilePath),
		EntrypointOverride: newNullString(app.EntrypointOverride),
		Type:               appTypePbToModel(app.Type),
		InternalDNS:        newNullString(app.InternalDns),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		CreationState:      models.CreationStateTypeCreateInfrastructureWait,
	}

	if err := model.Insert(s.db); err != nil {
		return nil, errors.Wrap(err, "inserting into db")
	}

	app.Id = int32(model.ID)

	// start a terraform job in the background
	go s.createAppInfrastructure(app)

	return app, nil
}

func (s *server) createAppInfrastructure(app *pb.Application) {
	setState := func(state pb.CreationState) {
		app.CreationState = state
		updateSQL := "UPDATE applications SET creation_state = $1 WHERE id = $2"
		if _, err := s.db.Exec(updateSQL, creationStateTypePbToModel(state), app.GetId()); err != nil {
			errors.Wrap(err, "updating applications table")
		}
	}

	do := func(f func() error) {
		if app.CreationState != pb.CreationState_FAILED {
			if err := f(); err != nil {
				log.Printf("app creation failed: %v", err)
				setState(pb.CreationState_FAILED)
			}
		}
	}

	switch app.GetCreationState() {
	case pb.CreationState_CREATE_INFRASTRUCTURE_WAIT:
		// run terraform apply on VPC config TODO(paulsmith):
		// bundle the terraform configs with the Soapbox app
		// and make them available in a well-known location
		terraformPath := filepath.Join("ops", "aws", "terraform")
		scriptsPath := filepath.Join(terraformPath, "scripts")

		slug := app.GetSlug()
		var networkDir, deploymentDir string

		do(func() error {
			log.Printf("generating terraform configuration - network")
			cmd := exec.Command("./init_app_tf.sh",
				"-a", slug,
				"-e", "test", // TODO(paulsmith): FIXME
				"-t", "network")
			cmd.Dir = scriptsPath
			var buf bytes.Buffer
			cmd.Stdout = &buf
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return errors.Wrap(err, "running init_app_tf.sh for network")
			}
			networkDir = strings.TrimSpace(buf.String())
			return nil
		})

		if networkDir != "" {
			defer os.RemoveAll(networkDir)
		}

		do(func() error {
			log.Printf("running terraform plan - network")
			cmd := exec.Command("terraform", "plan",
				"-var", "application_name="+slug,
				"-var", "environment=test", // TODO(paulsmith): FIXME
				"-no-color")
			cmd.Dir = networkDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		do(func() error {
			log.Printf("running terraform apply - network")
			cmd := exec.Command("terraform", "apply",
				"-var", "application_name="+slug,
				"-var", "environment=test", // TODO(paulsmith): FIXME
				"-no-color")
			cmd.Dir = networkDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		do(func() error {
			log.Printf("generating terraform configuration - deployment")
			cmd := exec.Command("./init_app_tf.sh",
				"-a", slug,
				"-e", "test", // TODO(paulsmith): FIXME
				"-t", "deployment")
			cmd.Dir = scriptsPath
			var buf bytes.Buffer
			cmd.Stdout = &buf
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return errors.Wrap(err, "running init_app_tf.sh for deployment")
			}
			deploymentDir = strings.TrimSpace(buf.String())
			return nil
		})

		if deploymentDir != "" {
			defer os.RemoveAll(deploymentDir)
		}

		do(func() error {
			cmd := exec.Command("terraform", "get")
			cmd.Dir = filepath.Join(deploymentDir, "asg")
			return cmd.Run()
		})

		do(func() error {
			log.Printf("running terraform plan - deployment")
			cmd := exec.Command("terraform", "plan",
				"-var", "application_name="+slug,
				"-var", "environment=test", // TODO(paulsmith): FIXME
				"-no-color")
			cmd.Dir = filepath.Join(deploymentDir, "asg")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		do(func() error {
			log.Printf("running terraform apply - deployment")
			cmd := exec.Command("terraform", "apply",
				"-var", "application_name="+slug,
				"-var", "environment=test", // TODO(paulsmith): FIXME
				"-no-color")
			cmd.Dir = filepath.Join(deploymentDir, "asg")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		do(func() error {
			setState(pb.CreationState_SUCCEEDED)
			log.Printf("done")
			if err := s.AddApplicationActivity(context.Background(), app.GetId(), app.GetUserId()); err != nil {
				return errors.Wrap(err, "error adding activity")
			}
			return nil
		})
	case pb.CreationState_SUCCEEDED:
		log.Printf("creation already succeeded, doing nothing")
	case pb.CreationState_FAILED:
		// TODO(paulsmith): advance this to
		// CREATE_INFRASTRUCTURE_WAIT state with some retry
		// logic like max attempts
		log.Printf("creation previously failed, should retry")
	}
}

type httpHead interface {
	Head(url string) (*http.Response, error)
}

func canAccessURL(client httpHead, url string) error {
	resp, err := client.Head(url)
	if err != nil {
		return errors.Wrapf(err, "couldn't make HTTP HEAD request to %s", url)
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return errors.Wrapf(err, "non-success HTTP status response from %s: %d", url, resp.StatusCode)
	}
	return nil
}

func (s *server) ListApplications(ctx context.Context, req *pb.ListApplicationRequest) (*pb.ListApplicationResponse, error) {
	const query = `SELECT id, name, description, created_at FROM applications WHERE user_id = $1 ORDER BY created_at ASC`
	rows, err := s.db.Query(query, req.UserId)
	if err != nil {
		return nil, errors.Wrap(err, "querying db for apps list")
	}

	var apps []*pb.Application

	for rows.Next() {
		app := &pb.Application{
			CreatedAt: new(gpb.Timestamp),
		}
		var createdAt time.Time
		dest := []interface{}{&app.Id, &app.Name, &app.Description, &createdAt}
		if err := rows.Scan(dest...); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}
		setPbTimestamp(app.CreatedAt, createdAt)
		apps = append(apps, app)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterating over db rows")
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

func creationStateTypeModelToPb(cst models.CreationStateType) pb.CreationState {
	switch cst {
	case models.CreationStateTypeCreateInfrastructureWait:
		return pb.CreationState_CREATE_INFRASTRUCTURE_WAIT
	case models.CreationStateTypeSucceeded:
		return pb.CreationState_SUCCEEDED
	case models.CreationStateTypeFailed:
		return pb.CreationState_FAILED
	default:
		panic("shouldn't get here")
	}
}

func creationStateTypePbToModel(cst pb.CreationState) models.CreationStateType {
	switch cst {
	case pb.CreationState_CREATE_INFRASTRUCTURE_WAIT:
		return models.CreationStateTypeCreateInfrastructureWait
	case pb.CreationState_SUCCEEDED:
		return models.CreationStateTypeSucceeded
	case pb.CreationState_FAILED:
		return models.CreationStateTypeFailed
	default:
		panic("shouldn't get here")
	}
}

func (s *server) GetApplication(ctx context.Context, req *pb.GetApplicationRequest) (*pb.Application, error) {
	model, err := models.ApplicationByID(s.db, int(req.Id))
	if err != nil {
		return nil, errors.Wrap(err, "getting application by ID from db")
	}

	app := &pb.Application{
		Id:        int32(model.ID),
		UserId:    int32(model.UserID),
		Name:      model.Name,
		Slug:      model.Slug,
		Type:      appTypeModelToPb(model.Type),
		CreatedAt: new(gpb.Timestamp),
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

	setPbTimestamp(app.CreatedAt, model.CreatedAt)

	app.CreationState = creationStateTypeModelToPb(model.CreationState)

	return app, nil
}

func setPbTimestamp(ts *gpb.Timestamp, t time.Time) {
	ts.Seconds = t.Unix()
}
