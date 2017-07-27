package api

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/adhocteam/soapbox/models"
	pb "github.com/adhocteam/soapbox/soapboxpb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type server struct {
	db         *sql.DB
	httpClient *http.Client
}

type state string

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

func nullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

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
				errors.Wrap(err, "app creation failed")
				setState(pb.CreationState_FAILED)
			}
		}
	}

	switch app.GetCreationState() {
	case pb.CreationState_CREATE_INFRASTRUCTURE_WAIT:
		// run terraform apply on VPC config TODO(paulsmith):
		// bundle the terraform configs with the Soapbox app
		// and make them available in a well-known location
		var currdir string
		do(func() error {
			var err error
			currdir, err = os.Getwd()
			return err
		})

		pathToTerraformConfig := filepath.Join("ops", "aws", "terraform")
		do(func() error {
			return os.Chdir(filepath.Join(pathToTerraformConfig, "network"))
		})

		slug := app.GetSlug()
		domain := fmt.Sprintf("%s.soapboxhosting.computer", slug)

		do(func() error {
			log.Printf("running terraform plan - network")
			// TODO(paulsmith): resolve issue around VPCs per app or per env
			cmd := exec.Command("terraform", "plan", "-var", "application_domain="+domain, "-var", "application_name="+slug, "-var", "environment=test", "-out=/dev/null", "-state=/dev/null", "-lock=false")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		do(func() error {
			log.Printf("running terraform apply - network")
			// TODO(paulsmith): resolve issue around VPCs per app or per env
			cmd := exec.Command("terraform", "apply", "-var", "application_domain="+domain, "-var", "application_name="+slug, "-var", "environment=test")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		do(func() error {
			return os.Chdir(filepath.Join("..", "deployment", "asg"))
		})

		do(func() error {
			return exec.Command("terraform", "get").Run()
		})

		do(func() error {
			log.Printf("running terraform plan - deployment")
			// TODO(paulsmith): resolve issue around VPCs per app or per env
			cmd := exec.Command("terraform", "plan", "-var", "application_domain="+domain, "-var", "application_name="+slug, "-var", "environment=test", "-out=/dev/null", "-state=/dev/null", "-lock=false")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		do(func() error {
			log.Printf("running terraform apply - deployment")
			// TODO(paulsmith): resolve issue around VPCs per app or per env
			cmd := exec.Command("terraform", "apply", "-var", "application_domain="+domain, "-var", "application_name="+slug, "-var", "environment=test")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})

		do(func() error {
			setState(pb.CreationState_SUCCEEDED)
			log.Printf("done")
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
		return nil, errors.Wrap(err, "querying db for apps list")
	}

	var apps []*pb.Application

	for rows.Next() {
		var a pb.Application
		dest := []interface{}{&a.Id, &a.Name, &a.Description, &a.CreatedAt}
		if err := rows.Scan(dest...); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}
		apps = append(apps, &a)
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

	app.CreationState = creationStateTypeModelToPb(model.CreationState)

	return app, nil
}

const timestampFormat = "2006-01-02T15:04:05"

func (s *server) ListEnvironments(ctx context.Context, req *pb.ListEnvironmentRequest) (*pb.ListEnvironmentResponse, error) {
	listSQL := "SELECT id, application_id, name, slug, vars, created_at FROM environments WHERE application_id = $1 ORDER BY id"
	rows, err := s.db.Query(listSQL, req.GetApplicationId())
	if err != nil {
		return nil, errors.Wrap(err, "querying db for environments")
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
			return nil, errors.Wrap(err, "scanning db row")
		}
		if err := json.Unmarshal(vars, &env.Vars); err != nil {
			return nil, errors.Wrap(err, "unmarshalling env vars JSON")
		}
		envs = append(envs, &env)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterating over db result set")
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
		return nil, errors.Wrap(err, "scanning db row")
	}
	if err := json.Unmarshal(vars, &env.Vars); err != nil {
		return nil, errors.Wrap(err, "unmarshalling env vars JSON")
	}
	return &env, nil
}

func (s *server) CreateEnvironment(ctx context.Context, req *pb.Environment) (*pb.Environment, error) {
	// TODO(paulsmith): can we even do this in XO??
	insertSQL := "INSERT INTO environments (application_id, name, slug, vars) VALUES ($1, $2, $3, $4) RETURNING id, created_at"

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req.Vars); err != nil {
		return nil, errors.Wrap(err, "encoding env vars as JSON")
	}

	args := []interface{}{
		req.GetApplicationId(),
		req.GetName(),
		slugify(req.GetName()),
		buf.String(),
	}

	var id int

	if err := s.db.QueryRow(insertSQL, args...).Scan(&id, &req.CreatedAt); err != nil {
		return nil, errors.Wrap(err, "inserting in to db")
	}

	req.Id = int32(id)

	return req, nil
}

func (s *server) DestroyEnvironment(ctx context.Context, req *pb.DestroyEnvironmentRequest) (*pb.Empty, error) {
	deleteSQL := "DELETE FROM environments WHERE id = $1"
	if _, err := s.db.Exec(deleteSQL, req.GetId()); err != nil {
		return nil, errors.Wrap(err, "deleting row from db")
	}
	return &pb.Empty{}, nil
}

func (s *server) CopyEnvironment(context.Context, *pb.CopyEnvironmentRequest) (*pb.Environment, error) {
	return nil, nil
}

type deploymentState struct {
	*pb.DeploymentState
}

func (c deploymentState) Scan(src interface{}) error {
	switch val := src.(type) {
	case []byte:
		return c.UnmarshalText(val)
	default:
		return fmt.Errorf("invalid DeploymentStateType: %#v", src)
	}
}

func (c deploymentState) UnmarshalText(b []byte) error {
	if state, exists := pb.DeploymentState_value[string(b)]; exists {
		*c.DeploymentState = pb.DeploymentState(state)
	} else if !exists {
		return fmt.Errorf("unknown value for deployment state: %v", string(b))
	}

	return nil
}

func (s *server) ListDeployments(ctx context.Context, req *pb.ListDeploymentRequest) (*pb.ListDeploymentResponse, error) {
	listSQL := "SELECT d.id, d.application_id, d.environment_id, d.committish, d.current_state, d.created_at, e.name FROM deployments d, environments e WHERE d.environment_id = e.id AND d.application_id = $1"
	rows, err := s.db.Query(listSQL, req.GetApplicationId())
	if err != nil {
		return nil, errors.Wrap(err, "querying db in ListDeployments")
	}
	var deployments []*pb.Deployment

	defer rows.Close()
	for rows.Next() {
		var d pb.Deployment
		d.Application = &pb.Application{}
		d.Env = &pb.Environment{}
		dest := []interface{}{
			&d.Id,
			&d.Application.Id,
			&d.Env.Id,
			&d.Committish,
			deploymentState{&d.State},
			&d.CreatedAt,
			&d.Env.Name,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, errors.Wrap(err, "scanning db row")
		}
		deployments = append(deployments, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iteration over result set")
	}
	res := &pb.ListDeploymentResponse{
		Deployments: deployments,
	}
	return res, nil
}

func (s *server) GetDeployment(ctx context.Context, req *pb.GetDeploymentRequest) (*pb.Deployment, error) {
	return nil, nil

}

func (s *server) StartDeployment(ctx context.Context, req *pb.Deployment) (*pb.StartDeploymentResponse, error) {
	req.State = pb.DeploymentState_DEPLOYMENT_ROLLOUT_WAIT
	query := `INSERT INTO deployments (application_id, environment_id, committish, current_state) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	appId := int(req.GetApplication().GetId())
	args := []interface{}{
		appId,
		req.GetEnv().GetId(),
		req.GetCommittish(),
		pb.DeploymentStateToString(req.GetState()),
	}
	dest := []interface{}{
		&req.Id,
		&req.CreatedAt,
	}
	if err := s.db.QueryRow(query, args...).Scan(dest...); err != nil {
		return nil, errors.Wrap(err, "inserting new row into db")
	}
	env := req.GetEnv()
	// TODO(paulsmith): hydrate fields for app and env
	app, err := models.ApplicationByID(s.db, appId)
	if err != nil {
		return nil, errors.Wrap(err, "getting application model from db")
	}
	req.Application.Name = app.Name
	req.Application.Description = nullString(app.Description)
	req.Application.GithubRepoUrl = nullString(app.GithubRepoURL)
	req.Application.Slug = app.Slug
	if err := s.db.QueryRow("SELECT slug FROM environments WHERE id = $1", env.Id).Scan(&env.Slug); err != nil {
		return nil, errors.Wrap(err, "querying for env slug")
	}
	req.Env = env
	go s.startDeployment(req)
	res := &pb.StartDeploymentResponse{
		Id: req.GetId(),
	}
	return res, nil
}

func (s *server) GetDeploymentStatus(ctx context.Context, req *pb.GetDeploymentStatusRequest) (*pb.GetDeploymentStatusResponse, error) {
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

func (s *server) TeardownDeployment(ctx context.Context, req *pb.TeardownDeploymentRequest) (*pb.Empty, error) {
	return nil, nil
}

var sha1Re = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)

const deployStateTagName = "deploystate"

func (s *server) startDeployment(dep *pb.Deployment) {
	setState := func(state pb.DeploymentState) {
		if dep.State == pb.DeploymentState_DEPLOYMENT_FAILED {
			return
		}
		dep.State = state
		updateSQL := "UPDATE deployments SET current_state = $1 WHERE id = $2"
		if _, err := s.db.Exec(updateSQL, pb.DeploymentStateToString(state), dep.GetId()); err != nil {
			log.Printf("updating deployments table: %v", err)
		}
	}

	do := func(f func() error) {
		if dep.State != pb.DeploymentState_DEPLOYMENT_FAILED {
			if err := f(); err != nil {
				log.Printf("error: %v", err)
				setState(pb.DeploymentState_DEPLOYMENT_FAILED)
			}
		}
	}

	doCmd := func(cmd string, args ...string) {
		do(func() error {
			out, err := exec.Command(cmd, args...).CombinedOutput()
			if err != nil {
				log.Printf("command %s %s:\n%s", cmd, args, out)
			}
			return err
		})
	}

	app := newAppFromProtoBuf(dep.GetApplication())

	do(func() error {
		var err error
		app.sess, err = session.NewSession()
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

	do(func() error { return os.Chdir(tempdir) })

	// clone the repo at the committish
	const appdir = "appdir"
	log.Printf("cloning repo")
	doCmd("git", "clone", app.githubRepoUrl, appdir)

	do(func() error {
		return os.Chdir(appdir)
	})

	log.Println("checking out committish")
	doCmd("git", "checkout", dep.GetCommittish())

	image := fmt.Sprintf("soapbox/%s:latest", app.slug)

	// build the docker image from the repo
	log.Printf("building docker image: %s", image)
	doCmd("docker", "build", "-t", image, ".")

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
	const soapboxImageBucket = "soapbox-app-images"
	// TODO(paulsmith): add committish to filename (and
	// update downstream code to match for it when
	// fetching on app instance)
	objectKey := fmt.Sprintf("%s/%s.tar.gz", app.slug, app.slug)
	do(func() error {
		return newS3Storage(app.sess).uploadFile(soapboxImageBucket, objectKey, filename)
	})

	setState(pb.DeploymentState_DEPLOYMENT_EVALUATE_WAIT)

	// start an ec2 instance, passing a user-data script which
	// installs the docker image and gets the container running
	userDataTmpl := `#!/bin/bash

set -xeuo pipefail

# log all script output
exec > >(tee /var/log/user-data.log) 2>&1

AWS=/usr/bin/aws
DOCKER=/usr/bin/docker

APP_NAME="{{.Slug}}"
PORT="{{.ListenPort}}"
RELEASE_BUCKET="{{.Bucket}}"
ENV="{{.Environment}}"
IMAGE="{{.Image}}"

# Retrieve the release from s3
$AWS s3 cp s3://$RELEASE_BUCKET/$APP_NAME/$APP_NAME.tar.gz /tmp/$APP_NAME.tar.gz

# Install the docker image
$DOCKER image load -i /tmp/$APP_NAME.tar.gz

# Set up the runit dirs
mkdir -p "/etc/sv/$APP_NAME"

# Logging configuration
mkdir -p "/etc/sv/$APP_NAME/log"
mkdir -p "/var/log/$APP_NAME"

# Create the logging run script
cat << EOF > /etc/sv/$APP_NAME/log/run
#!/bin/sh
exec svlogd -tt /var/log/$APP_NAME
EOF

# Mark the log/run file executable
chmod +x /etc/sv/$APP_NAME/log/run

# Create the run script for the app
cat << EOF > /etc/sv/$APP_NAME/run
#!/bin/bash
exec 2>&1 chpst -e /etc/sv/$APP_NAME/env $DOCKER run --rm --name $APP_NAME-run -p 9090:$PORT "$IMAGE"
EOF

# Mark the run file executable
chmod +x /etc/sv/$APP_NAME/run

# Create a link from /etc/service/$APP_NAME -> /etc/sv/$APP_NAME
ln -s /etc/sv/$APP_NAME /etc/service/$APP_NAME
`
	var tmpl *template.Template
	do(func() error {
		var err error
		tmpl, err = template.New("user-data.tmpl").Parse(userDataTmpl)
		return err
	})
	var userData bytes.Buffer
	do(func() error {
		return tmpl.Execute(&userData, struct {
			Slug        string
			ListenPort  int
			Bucket      string
			Environment string
			Image       string
		}{
			app.slug,
			// TODO(paulsmith): un-hardcode
			8080,
			soapboxImageBucket,
			// TODO(paulsmith): unused in user-data script atm
			"",
			image,
		})
	})

	env := newEnvFromProtoBuf(dep.GetEnv())

	committish := dep.GetCommittish()
	if sha1Re.MatchString(committish) {
		committish = committish[:7]
	}

	var securityGroupId string
	do(func() error {
		var err error
		securityGroupId, err = app.getAppSecurityGroupId(env)
		return err
	})

	var launchConfig string
	do(func() error {
		var err error
		launchConfig, err = createLaunchConfig(app, env, committish, securityGroupId, time.Now(), userData.String())
		return err
	})

	log.Printf("created launch config: %s", launchConfig)

	var blueASG, greenASG *autoScalingGroup
	do(func() error {
		var err error
		blueASG, greenASG, err = app.blueGreenASGs(env)
		return err
	})

	log.Printf("blue ASG is currently: %s", blueASG.name)
	log.Printf("green ASG is currently: %s", greenASG.name)

	const nAZs = 2 // number of availability zones

	log.Printf("updating blue ASG with new launch config")
	do(func() error {
		return blueASG.updateLaunchConfig(launchConfig, nAZs)
	})

	log.Printf("tagging blue ASG with release info")
	do(func() error {
		return blueASG.updateTags([]tag{{key: "release", value: committish}})
	})

	log.Printf("waiting for blue ASG instances to be ready")
	do(func() error { return blueASG.waitUntilInstancesReady(nAZs) })
	log.Printf("blue ASG instances ready")

	var target *targetGroup
	do(func() error {
		var err error
		target, err = greenASG.getTargetGroup()
		return err
	})

	log.Printf("attaching blue ASG to load balancer")
	do(func() error {
		return blueASG.attachToLBTargetGroup(target.arn)
	})

	log.Printf("waiting for blue ASG instances to pass health checks in load balancer")
	do(func() error {
		return target.waitUntilInstancesReady(blueASG)
	})

	setState(pb.DeploymentState_DEPLOYMENT_ROLL_FORWARD)

	log.Printf("detaching (stale) green ASG from load balancer")
	do(func() error {
		return greenASG.detachFromLBTargetGroup(target.arn)
	})

	log.Printf("scaling down (stale) green ASG")
	do(func() error {
		return greenASG.resize(0, 0, 0)
	})

	// TODO(paulsmith): there is a race condition because we can't
	// update the tags atomically, so a reader might see both
	// groups as green, or blue, or some indeterminate combination
	// ... risk is pretty low ATM but we should address this
	// somehow later.
	log.Printf("swapping blue/green pointers")
	do(func() error {
		return greenASG.updateTags([]tag{{key: deployStateTagName, value: "blue"}})
	})
	do(func() error {
		return blueASG.updateTags([]tag{{key: deployStateTagName, value: "green"}})
	})

	log.Printf("done")

	// TODO(paulsmith): health check

	setState(pb.DeploymentState_DEPLOYMENT_SUCCEEDED)
}

type targetGroup struct {
	svc *elbv2.ELBV2
	arn string
}

func (g *targetGroup) waitUntilInstancesReady(asg *autoScalingGroup) error {
	instances, err := asg.getInstances()
	if err != nil {
		return fmt.Errorf("getting ASG's instances: %v", err)
	}
	targets := make([]*elbv2.TargetDescription, len(instances))
	for i, inst := range instances {
		targets[i] = &elbv2.TargetDescription{
			Id: inst.InstanceId,
		}
	}
	input := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(g.arn),
		Targets:        targets,
	}
	deadline := time.Now().Add(5 * time.Minute)
	for {
		res, err := g.svc.DescribeTargetHealth(input)
		if err != nil {
			return fmt.Errorf("describing target group health: %v", err)
		}
		allHealthy := true
		for _, health := range res.TargetHealthDescriptions {
			// TargetHealthStateEnum:
			// - initial
			// - healthy
			// - unhealthy
			// - unused
			// - draining
			if *health.TargetHealth.State != "healthy" {
				allHealthy = false
				break
			}
		}
		if allHealthy {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for target group instances to be healthy")
		}
		time.Sleep(5 * time.Second)
	}
}

type tag struct {
	key               string
	value             string
	propagateAtLaunch bool
}

func (t tag) autoscaling(name string) *autoscaling.Tag {
	return &autoscaling.Tag{
		Key:               aws.String(t.key),
		ResourceId:        aws.String(name),
		ResourceType:      aws.String("auto-scaling-group"),
		Value:             aws.String(t.value),
		PropagateAtLaunch: aws.Bool(t.propagateAtLaunch),
	}
}

type autoScalingGroup struct {
	sess *session.Session
	svc  *autoscaling.AutoScaling
	name string
}

func (g *autoScalingGroup) updateTags(tags []tag) error {
	input := &autoscaling.CreateOrUpdateTagsInput{
		Tags: make([]*autoscaling.Tag, len(tags)),
	}
	for i, tag := range tags {
		input.Tags[i] = tag.autoscaling(g.name)
	}
	if _, err := g.svc.CreateOrUpdateTags(input); err != nil {
		return errors.Wrap(err, "updating ASG tags: ")
	}
	return nil
}

func (g *autoScalingGroup) updateLaunchConfig(lcName string, size int) error {
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName:    aws.String(g.name),
		LaunchConfigurationName: aws.String(lcName),
		DesiredCapacity:         aws.Int64(int64(size)),
		MaxSize:                 aws.Int64(int64(size)),
		MinSize:                 aws.Int64(int64(size)),
	}
	_, err := g.svc.UpdateAutoScalingGroup(input)
	return err
}

func (g *autoScalingGroup) attachToLBTargetGroup(targetGroupARN string) error {
	input := &autoscaling.AttachLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: aws.String(g.name),
		TargetGroupARNs: []*string{
			aws.String(targetGroupARN),
		},
	}
	_, err := g.svc.AttachLoadBalancerTargetGroups(input)
	return err
}

func (g *autoScalingGroup) getTargetGroup() (*targetGroup, error) {
	input := &autoscaling.DescribeLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: aws.String(g.name),
	}
	res, err := g.svc.DescribeLoadBalancerTargetGroups(input)
	if err != nil {
		return nil, err
	}
	group := res.LoadBalancerTargetGroups[0]
	target := &targetGroup{
		svc: elbv2.New(g.sess),
		arn: *group.LoadBalancerTargetGroupARN,
	}
	return target, nil
}

func (g *autoScalingGroup) detachFromLBTargetGroup(targetGroupARN string) error {
	input := &autoscaling.DetachLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: aws.String(g.name),
		TargetGroupARNs: []*string{
			aws.String(targetGroupARN),
		},
	}
	_, err := g.svc.DetachLoadBalancerTargetGroups(input)
	return err
}

func (g *autoScalingGroup) resize(min, max, desired int) error {
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(g.name),
		DesiredCapacity:      aws.Int64(int64(desired)),
		MaxSize:              aws.Int64(int64(max)),
		MinSize:              aws.Int64(int64(min)),
	}
	_, err := g.svc.UpdateAutoScalingGroup(input)
	return err
}

func (g *autoScalingGroup) getInstances() ([]*autoscaling.Instance, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(g.name)},
	}
	res, err := g.svc.DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, fmt.Errorf("describing ASG: %v", err)
	}
	group := res.AutoScalingGroups[0]
	return group.Instances, nil
}

func createLaunchConfig(app *application, env *environment, committish string, securityGroupId string, t time.Time, userData string) (string, error) {
	name := fmt.Sprintf("%s-%s-%s-%d", app.slug, env.slug, committish, t.Unix())

	// TODO(paulsmith): get from Soapbox platform config
	const iamInstanceProfile = "soapbox-app"
	const appAmi = "ami-ef545cf9"
	const instanceType = "t2.micro"
	const keyName = "soapbox-app"

	input := &autoscaling.CreateLaunchConfigurationInput{
		IamInstanceProfile:      aws.String(iamInstanceProfile),
		ImageId:                 aws.String(appAmi),
		InstanceType:            aws.String(instanceType),
		KeyName:                 aws.String(keyName),
		LaunchConfigurationName: aws.String(name),
		SecurityGroups:          []*string{aws.String(securityGroupId)},
		UserData:                aws.String(base64.StdEncoding.EncodeToString([]byte(userData))),
	}

	svc := autoscaling.New(app.sess)
	_, err := svc.CreateLaunchConfiguration(input)
	if err != nil {
		return "", fmt.Errorf("creating launch config: %v", err)
	}

	return name, nil
}

type s3storage struct {
	svc *s3.S3
}

func newS3Storage(sess *session.Session) *s3storage {
	return &s3storage{svc: s3.New(sess)}
}

func (s *s3storage) uploadFile(bucket string, key string, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening file %s: %v", filename, err)
	}
	defer f.Close()
	input := &s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	_, err = s.svc.PutObject(input)
	return err
}

type application struct {
	name          string
	slug          string
	githubRepoUrl string

	sess *session.Session
}

func newAppFromProtoBuf(appPb *pb.Application) *application {
	return &application{
		name:          appPb.GetName(),
		slug:          appPb.GetSlug(),
		githubRepoUrl: appPb.GetGithubRepoUrl(),
	}
}

type environment struct {
	name string
	slug string
}

func newEnvFromProtoBuf(envPb *pb.Environment) *environment {
	return &environment{
		name: envPb.GetName(),
		slug: envPb.GetSlug(),
	}
}

func (a *application) blueGreenASGs(env *environment) (blue *autoScalingGroup, green *autoScalingGroup, err error) {
	// get a list of all ASGs and iterate over until find
	// "deploystate" tags for our app and environment

	svc := autoscaling.New(a.sess)
	asgs, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return nil, nil, fmt.Errorf("describing ASGs: %v", err)
	}

	type mapkey struct {
		app         string
		env         string
		deploystate string
	}

	groups := make(map[mapkey]*autoscaling.Group)

	for _, asg := range asgs.AutoScalingGroups {
		var key mapkey
		for _, tag := range asg.Tags {
			switch *tag.Key {
			case "app":
				key.app = *tag.Value
			case "env":
				key.env = *tag.Value
			case deployStateTagName:
				key.deploystate = *tag.Value
			}
		}
		groups[key] = asg
	}

	blueASG, ok := groups[mapkey{app: a.slug, env: env.slug, deploystate: "blue"}]
	if !ok {
		return nil, nil, fmt.Errorf("could not find blue ASG in %s environment", env.slug)
	}

	greenASG, ok := groups[mapkey{app: a.slug, env: env.slug, deploystate: "green"}]
	if !ok {
		return nil, nil, fmt.Errorf("could not find green ASG in %s environment", env.slug)
	}

	blue = &autoScalingGroup{
		sess: a.sess,
		svc:  svc,
		name: *blueASG.AutoScalingGroupName,
	}

	green = &autoScalingGroup{
		sess: a.sess,
		svc:  svc,
		name: *greenASG.AutoScalingGroupName,
	}

	return blue, green, nil
}

func (a *application) getAppSecurityGroupId(env *environment) (string, error) {
	sgname := fmt.Sprintf("%s: %s application subnet security group", a.slug, env.slug)
	svc := ec2.New(a.sess)
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []*string{aws.String(sgname)},
			},
		},
	}

	res, err := svc.DescribeSecurityGroups(input)
	if err != nil {
		return "", err
	}
	sg := res.SecurityGroups[0]
	return *sg.GroupId, nil
}

// wait for instances to be marked in-service in the ASG lifecycle
func (g *autoScalingGroup) waitUntilInstancesReady(n int) error {
	deadline := time.Now().Add(5 * time.Minute)
	for {
		count, err := inService(g.svc, g.name)
		if err != nil {
			return err
		}
		if count == n {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for ASG instances to be ready)")
		}
		time.Sleep(5 * time.Second)
	}
}

func inService(svc *autoscaling.AutoScaling, name string) (int, error) {
	out, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(name)},
	})
	if err != nil {
		return 0, err
	}

	count := 0
	group := out.AutoScalingGroups[0]
	for _, inst := range group.Instances {
		if *inst.LifecycleState == "InService" {
			count++
		}
	}
	return count, nil
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
