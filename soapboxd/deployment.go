package soapboxd

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/adhocteam/soapbox/models"
	pb "github.com/adhocteam/soapbox/proto"
	gpb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

//SoapboxImageBucket is the location within the blob store to hold build images
const SoapboxImageBucket = "soapbox-app-images"

func (s *Server) ListDeployments(ctx context.Context, req *pb.ListDeploymentRequest) (*pb.ListDeploymentResponse, error) {
	listSQL := "SELECT d.id, d.application_id, d.environment_id, d.committish, d.current_state, d.created_at, e.name FROM deployments d, environments e WHERE d.environment_id = e.id AND d.application_id = $1"
	rows, err := s.db.Query(listSQL, req.GetApplicationId())
	if err != nil {
		return nil, errors.Wrap(err, "querying db")
	}
	var deployments []*pb.Deployment
	for rows.Next() {
		d := &pb.Deployment{
			Application: &pb.Application{},
			Env:         &pb.Environment{},
			CreatedAt:   new(gpb.Timestamp),
		}
		var createdAt time.Time
		dest := []interface{}{
			&d.Id,
			&d.Application.Id,
			&d.Env.Id,
			&d.Committish,
			&d.State,
			&createdAt,
			&d.Env.Name,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}
		setPbTimestamp(d.CreatedAt, createdAt)
		deployments = append(deployments, d)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iteration over result set")
	}

	// environments
	envReq := &pb.ListEnvironmentRequest{
		ApplicationId: req.GetApplicationId(),
	}
	envRes, err := s.ListEnvironments(ctx, envReq)
	if err != nil {
		return nil, errors.Wrap(err, "getting environments")
	}
	byID := make(map[int32]*pb.Environment)
	for _, env := range envRes.Environments {
		byID[env.GetId()] = env
	}
	for _, d := range deployments {
		d.Env = byID[d.GetEnv().GetId()]
	}

	res := &pb.ListDeploymentResponse{
		Deployments: deployments,
	}
	return res, nil
}

func (s *Server) GetDeployment(ctx context.Context, req *pb.GetDeploymentRequest) (*pb.Deployment, error) {
	return nil, nil
}

// GetLatestDeployment gets latest deployment for an application environment.
func (s *Server) GetLatestDeployment(ctx context.Context, req *pb.GetLatestDeploymentRequest) (*pb.Deployment, error) {
	appID := req.GetApplicationId()
	envID := req.GetEnvironmentId()
	query := `
SELECT id, committish, current_state, created_at
FROM deployments
WHERE application_id = $1 AND environment_id = $2
ORDER BY id DESC
LIMIT 1
`
	var createdAt time.Time
	dep := &pb.Deployment{
		CreatedAt: new(gpb.Timestamp),
	}
	dep.Application = &pb.Application{Id: appID}
	dep.Env = &pb.Environment{Id: envID}
	dest := []interface{}{
		&dep.Id,
		&dep.Committish,
		&dep.State,
		&createdAt,
	}
	if err := s.db.QueryRow(query, appID, envID).Scan(dest...); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "no deployment found for application environment")
		}
		return nil, errors.Wrap(err, "querying deployments table")
	}
	setPbTimestamp(dep.CreatedAt, createdAt)
	return dep, nil
}

func (s *Server) StartDeployment(ctx context.Context, req *pb.Deployment) (*pb.StartDeploymentResponse, error) {
	req.State = "rollout-wait"
	query := `INSERT INTO deployments (application_id, environment_id, committish, current_state) VALUES ($1, $2, $3, $4) RETURNING id`
	appID := int(req.GetApplication().GetId())
	args := []interface{}{
		appID,
		req.GetEnv().GetId(),
		req.GetCommittish(),
		req.GetState(),
	}
	dest := []interface{}{
		&req.Id,
	}
	if err := s.db.QueryRow(query, args...).Scan(dest...); err != nil {
		return nil, errors.Wrap(err, "inserting new row into db")
	}
	// TODO(paulsmith): hydrate fields for app and env
	app, err := models.ApplicationByID(s.db, appID)
	if err != nil {
		return nil, errors.Wrap(err, "getting application model from db")
	}
	req.Application.Name = app.Name
	req.Application.Description = nullString(app.Description)
	req.Application.GithubRepoUrl = nullString(app.GithubRepoURL)
	req.Application.Slug = app.Slug
	req.Application.Id = int32(app.ID)
	req.Application.UserId = int32(app.UserID)

	envReq := pb.GetEnvironmentRequest{Id: req.GetEnv().GetId()}
	env, err := s.GetEnvironment(ctx, &envReq)
	if err != nil {
		return nil, errors.Wrap(err, "getting environment")
	}
	req.Env = env

	go s.startDeployment(req)
	res := &pb.StartDeploymentResponse{
		Id: req.GetId(),
	}

	if err := s.AddDeploymentActivity(context.Background(), pb.ActivityType_DEPLOYMENT_STARTED, req); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) GetDeploymentStatus(ctx context.Context, req *pb.GetDeploymentStatusRequest) (*pb.GetDeploymentStatusResponse, error) {
	var state string
	query := `SELECT current_state FROM deployments WHERE id = $1`
	if err := s.db.QueryRow(query, req.GetId()).Scan(&state); err != nil {
		return nil, errors.Wrap(err, "querying db for deploy state")
	}
	res := &pb.GetDeploymentStatusResponse{
		State: state,
	}
	return res, nil
}

func (s *Server) TeardownDeployment(ctx context.Context, req *pb.TeardownDeploymentRequest) (*pb.Empty, error) {
	return nil, nil
}

var sha1Re = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)

func (s *Server) startDeployment(dep *pb.Deployment) {
	setState := func(state string) {
		if dep.State == "failed" {
			return
		}
		dep.State = state
		updateSQL := "UPDATE deployments SET current_state = $1 WHERE id = $2"
		if _, err := s.db.Exec(updateSQL, state, dep.GetId()); err != nil {
			log.Printf("updating deployments table: %v", err)
		}
	}

	do := func(f func() error) {
		if dep.State != "failed" {
			if err := f(); err != nil {
				setState("failed")
				if err := s.AddDeploymentActivity(context.Background(), pb.ActivityType_DEPLOYMENT_FAILURE, dep); err != nil {
					log.Printf("adding deployment activity: %v", err)
				}
				log.Println(err)
			}
		}
	}

	doCmd := func(cmd *exec.Cmd) {
		do(func() error {
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("command %s %v:\n%s", cmd.Path, cmd.Args, out)
			}
			return err
		})
	}

	app := newAppFromProtoBuf(dep.GetApplication())
	env := newEnvFromProtoBuf(dep.GetEnv())

	var config *pb.Configuration
	do(func() error {
		var err error
		config, err = s.GetLatestConfiguration(context.TODO(), &pb.GetLatestConfigurationRequest{
			EnvironmentId: env.ID,
		})
		return err
	})

	// get a temp dir to work in
	var tempdir string
	do(func() error {
		var err error
		tempdir, err = ioutil.TempDir("", "sandbox")
		return err
	})

	if tempdir != "" {
		defer os.RemoveAll(tempdir)
	}

	// clone the repo at the committish
	const appdir = "appdir"
	cmd := exec.Command("git", "clone", app.GithubRepoURL, appdir)
	cmd.Dir = tempdir
	log.Println("cloning repo")
	doCmd(cmd)

	committish := dep.GetCommittish()
	cmd = exec.Command("git", "checkout", committish)
	cmd.Dir = filepath.Join(tempdir, appdir)
	log.Println("checking out committish")
	doCmd(cmd)

	if sha1Re.MatchString(committish) {
		// use short committish from here on out
		committish = committish[:7]
	}

	// Save the short committish into app for use during deployment
	app.Committish = committish

	image := fmt.Sprintf("soapbox/%s:%s", app.Slug, committish)

	// build the docker image from the repo
	cmd = exec.Command("docker", "build", "-t", image, ".")
	cmd.Dir = filepath.Join(tempdir, appdir)
	log.Printf("building docker image: %s", image)
	doCmd(cmd)

	// export the docker image to a file
	log.Printf("saving docker image %s to file", image)
	var filename string
	do(func() error {
		var err error
		filename, err = exportDockerImageToFile(tempdir, image)
		return err
	})

	// upload docker image to S3 bucket
	log.Println("upload docker image to S3")
	objectKey := fmt.Sprintf("%s/%s-%s.tar.gz", app.Slug, app.Slug, committish)
	do(func() error {
		return s.objectStore.UploadFile(SoapboxImageBucket, objectKey, filename)
	})

	setState("evaluate-wait")

	do(func() error {
		return s.deployer.Deploy(app, env, config)
	})

	setState("rollforward")

	do(func() error {
		return s.deployer.Rollforward(app, env)
	})

	// Should always run to either clean up failed blue nodes or outdated ones
	s.deployer.Cleanup(app, env)

	log.Printf("done")

	// TODO(paulsmith): health check?

	if dep.State != "failed" {
		setState("success")
		if err := s.AddDeploymentActivity(context.Background(), pb.ActivityType_DEPLOYMENT_SUCCESS, dep); err != nil {
			log.Println("error adding deployment activity", err)
		}
	}
}

// Application is a minimal version of the information about an app on the platform
type Application struct {
	Name          string
	Slug          string
	GithubRepoURL string
	Committish    string
}

func newAppFromProtoBuf(appPb *pb.Application) Application {
	return Application{
		Name:          appPb.GetName(),
		Slug:          appPb.GetSlug(),
		GithubRepoURL: appPb.GetGithubRepoUrl(),
	}
}

// Environment is a minimal version of the information about an environment into which an application can be deployed
type Environment struct {
	ID   int32
	Name string
	Slug string
}

func newEnvFromProtoBuf(envPb *pb.Environment) Environment {
	return Environment{
		ID:   envPb.GetId(),
		Name: envPb.GetName(),
		Slug: envPb.GetSlug(),
	}
}

func exportDockerImageToFile(dir string, image string) (string, error) {
	filename := fmt.Sprintf("docker-export-%d.tar.gz", time.Now().Unix())
	path := filepath.Join(dir, filename)

	ds := exec.Command("docker", "save", image)
	gzip := exec.Command("gzip")

	var buf bytes.Buffer

	pr, pw := io.Pipe()
	ds.Stdout = pw
	gzip.Stdin = pr
	ds.Stderr = &buf
	gzip.Stderr = &buf

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzip.Stdout = f

	ds.Start()
	gzip.Start()

	var dockerSaveErr, gzipErr error

	go func() {
		dockerSaveErr = ds.Wait()
		pw.Close()
	}()

	gzipErr = gzip.Wait()

	if dockerSaveErr != nil || gzipErr != nil {
		return "", fmt.Errorf("docker save / gzip pipeline: %v %v %s", dockerSaveErr, gzipErr, buf.String())
	}

	return path, nil
}
