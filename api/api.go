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
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/adhocteam/soapbox/models"
	pb "github.com/adhocteam/soapbox/soapboxpb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/net/context"
)

type server struct {
	db         *sql.DB
	httpClient *http.Client
	jobs       map[int32]*deploymentStatus
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

func (s *server) ListDeployments(ctx context.Context, req *pb.ListDeploymentRequest) (*pb.ListDeploymentResponse, error) {
	listSQL := "SELECT d.id, d.application_id, d.environment_id, d.committish, d.current_state, d.created_at, e.name FROM deployments d, environments e WHERE d.environment_id = e.id AND d.application_id = $1"
	rows, err := s.db.Query(listSQL, req.GetApplicationId())
	if err != nil {
		return nil, fmt.Errorf("querying db: %v", err)
	}
	var deployments []*pb.Deployment
	for rows.Next() {
		var d pb.Deployment
		d.Application = &pb.Application{}
		d.Env = &pb.Environment{}
		dest := []interface{}{
			&d.Id,
			&d.Application.Id,
			&d.Env.Id,
			&d.Committish,
			&d.State,
			&d.CreatedAt,
			&d.Env.Name,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("scanning db row: %v", err)
		}
		deployments = append(deployments, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration over result set: %v", err)
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
	insertSQL := "INSERT INTO deployments (application_id, environment_id, committish) VALUES ($1, $2, $3) RETURNING id, current_state, created_at"
	args := []interface{}{
		req.GetApplication().Id,
		req.GetEnv().Id,
		req.GetCommittish(),
	}
	dest := []interface{}{
		&req.Id,
		&req.State,
		&req.CreatedAt,
	}
	if err := s.db.QueryRow(insertSQL, args...).Scan(dest...); err != nil {
		return nil, fmt.Errorf("inserting new row into db: %v", err)
	}
	env := req.GetEnv()
	// TODO(paulsmith): hydrate fields for app and env
	if err := s.db.QueryRow("SELECT slug FROM environments WHERE id = $1", env.Id).Scan(&env.Slug); err != nil {
		return nil, fmt.Errorf("querying for env slug: %v", err)
	}
	req.Env = env
	go s.startDeployment(req)
	res := &pb.StartDeploymentResponse{
		Id: req.GetId(),
	}
	return res, nil
}

func (s *server) GetDeploymentStatus(ctx context.Context, req *pb.GetDeploymentStatusRequest) (*pb.GetDeploymentStatusResponse, error) {
	if s.jobs == nil {
		s.jobs = make(map[int32]*deploymentStatus)
	}
	status, ok := s.jobs[req.GetId()]
	if ok {
		status.Lock()
		state := status.currentState
		status.Unlock()
		res := &pb.GetDeploymentStatusResponse{
			State: string(state),
		}
		return res, nil
	}
	return nil, fmt.Errorf("unknown deployment %d", req.GetId())
}

func (s *server) TeardownDeployment(ctx context.Context, req *pb.TeardownDeploymentRequest) (*pb.Empty, error) {
	return nil, nil
}

type deploymentStatus struct {
	currentState state
	sync.Mutex
}

func (s *server) startDeployment(req *pb.Deployment) {
	status := deploymentStatus{
		currentState: "rollout-wait",
	}
	setState := func(newState state) {
		status.Lock()
		defer status.Unlock()
		status.currentState = newState
	}
	if s.jobs == nil {
		s.jobs = make(map[int32]*deploymentStatus)
	}
	s.jobs[req.GetId()] = &status

	// get a temp dir to work in
	tempdir, err := ioutil.TempDir("", "sandbox")
	log.Printf("temp dir: %s", tempdir)
	if err != nil {
		log.Printf("creating temp dir: %v", err)
		setState("failed")
		return
	}
	defer os.RemoveAll(tempdir)

	if err := os.Chdir(tempdir); err != nil {
		log.Printf("changing to temp dir: %v", err)
		setState("failed")
		return
	}

	// clone the repo at the committish
	const appdir = "appdir"
	appId := int(req.GetApplication().GetId())
	app, err := models.ApplicationByID(s.db, appId)
	if err != nil {
		log.Printf("getting application by ID from db: %v", err)
		setState("failed")
		return
	}

	log.Println("cloning repo")
	out, err := exec.Command("git", "clone", app.GithubRepoURL.String, appdir).CombinedOutput()
	if err != nil {
		log.Printf("cloning repo: %v %s", err, string(out))
		setState("failed")
	}

	if err := os.Chdir(appdir); err != nil {
		log.Printf("changing to app dir: %v", err)
		setState("failed")
		return
	}

	log.Println("checking out committish")
	out, err = exec.Command("git", "checkout", req.GetCommittish()).CombinedOutput()
	if err != nil {
		log.Printf("checking out committish: %v %s", err, string(out))
		setState("failed")
		return
	}

	slug := app.Slug
	image := fmt.Sprintf("soapbox/%s:latest", slug)

	// build the docker image from the repo
	log.Printf("building docker image: %s", image)
	out, err = exec.Command("docker", "build", "-t", image, ".").CombinedOutput()
	if err != nil {
		log.Printf("building docker image: %v %s", err, string(out))
		setState("failed")
		return
	}

	// export the docker image to a file
	filename := fmt.Sprintf("%s.tar.gz", slug)
	log.Printf("saving docker image %s to file: %s", image, filename)
	ds := exec.Command("docker", "save", image)
	gzip := exec.Command("gzip")

	var buf bytes.Buffer

	pr, pw := io.Pipe()
	ds.Stdout = pw
	gzip.Stdin = pr
	ds.Stderr = &buf
	gzip.Stderr = &buf

	f, err := os.Create(filename)
	if err != nil {
		log.Printf("opening file: %v", err)
		setState("failed")
		return
	}
	gzip.Stdout = f

	if err := ds.Start(); err != nil {
		log.Printf("starting docker save: %v", err)
		setState("failed")
		return
	}
	if err := gzip.Start(); err != nil {
		log.Printf("starting gzip: %v", err)
		setState("failed")
		return
	}

	var dockerSaveErr, gzipErr error

	go func() {
		dockerSaveErr = ds.Wait()
		pw.Close()
	}()

	gzipErr = gzip.Wait()

	if dockerSaveErr != nil || gzipErr != nil {
		log.Printf("docker save / gzip pipeline: %v %v %s", dockerSaveErr, gzipErr, buf.String())
		setState("failed")
		return
	}

	f.Close()

	// upload docker image to S3 bucket
	log.Println("upload docker image to S3")
	sess, err := session.NewSession()
	if err != nil {
		log.Printf("new AWS session: %v", err)
		setState("failed")
		return
	}
	svc := s3.New(sess)
	f, err = os.Open(filename)
	if err != nil {
		log.Printf("opening file for reading: %v", err)
		setState("failed")
		return
	}
	const soapboxImageBucket = "soapbox-app-images"
	objectKey := fmt.Sprintf("%s/%s.tar.gz", slug, slug)
	input := &s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(soapboxImageBucket),
		Key:    aws.String(objectKey),
	}
	if _, err = svc.PutObject(input); err != nil {
		log.Printf("putting object to S3: %v", err)
		setState("failed")
		return
	}

	setState("evaluate-wait")

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

# TODO: Create env var dir
mkdir -p "/etc/sv/$APP_NAME/env"

# TODO: Fetch env vars to above dir

# TODO: logging configuration
#mkdir -p "/etc/sv/$APP_NAME/log"
#mkdir -p "/var/log/$APP_NAME"

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
	tmpl, err := template.New("user-data.tmpl").Parse(userDataTmpl)
	if err != nil {
		log.Printf("parsing user data template: %v", err)
		setState("failed")
		return
	}
	var userData bytes.Buffer
	if err := tmpl.Execute(&userData, struct {
		Slug        string
		ListenPort  int
		Bucket      string
		Environment string
		Image       string
	}{
		slug,
		// TODO(paulsmith): un-hardcode
		8080,
		soapboxImageBucket,
		// TODO(paulsmith): unused in user-data script atm
		"",
		image,
	}); err != nil {
		log.Printf("executing user-data template: %v", err)
		setState("failed")
		return
	}

	// TODO(paulsmith): get from Soapbox platform config
	const iamInstanceProfile = "soapbox-app"
	const appAmi = "ami-ef545cf9"
	const instanceType = "t2.micro"
	const keyName = "soapbox-app"
	// TODO(paulsmith): get from AWS SDK
	securityGroupIds := []*string{aws.String("sg-436f1332")}
	const subnetId = "subnet-118c193d"

	asSvc := autoscaling.New(sess)
	committish := req.GetCommittish()
	sha1Re := regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
	if sha1Re.MatchString(committish) {
		committish = committish[:8]
	}
	lcName := fmt.Sprintf("%s-%s-%s-%d", slug, req.GetEnv().Slug, committish, time.Now().Unix())
	lcInput := &autoscaling.CreateLaunchConfigurationInput{
		IamInstanceProfile:      aws.String(iamInstanceProfile),
		ImageId:                 aws.String(appAmi),
		InstanceType:            aws.String(instanceType),
		KeyName:                 aws.String(keyName),
		LaunchConfigurationName: aws.String(lcName),
		SecurityGroups:          securityGroupIds,
		UserData:                aws.String(base64.StdEncoding.EncodeToString(userData.Bytes())),
	}
	log.Printf("creating launch config")
	if _, err := asSvc.CreateLaunchConfiguration(lcInput); err != nil {
		log.Printf("creating launch config: %v", err)
		setState("failed")
		return
	}
	log.Printf("created launch config: %s", lcName)

	const deployStateTagName = "deploystate"
	// get a list of all ASGs and iterate over until find "deploystate" tag
	asgs, err := asSvc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		log.Printf("describe ASGs: %v", err)
		setState("failed")
		return
	}

	var blueAsgName, greenAsgName string
	for _, asg := range asgs.AutoScalingGroups {
		for _, tag := range asg.Tags {
			if *tag.Key == deployStateTagName {
				switch *tag.Value {
				case "blue":
					blueAsgName = *asg.AutoScalingGroupName
				case "green":
					greenAsgName = *asg.AutoScalingGroupName
				}
			}
		}
	}
	if blueAsgName == "" || greenAsgName == "" {
		log.Printf("couldn't find blue/green ASGs; check %s tags", deployStateTagName)
		setState("failed")
		return
	}
	log.Printf("blue ASG is currently: %s", blueAsgName)
	log.Printf("green ASG is currently: %s", greenAsgName)

	const nAZs = 2 // number of availability zones

	blueAsgInput := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName:    aws.String(blueAsgName),
		LaunchConfigurationName: aws.String(lcName),
		DesiredCapacity:         aws.Int64(nAZs),
		MaxSize:                 aws.Int64(nAZs),
		MinSize:                 aws.Int64(nAZs),
	}
	log.Printf("updating blue ASG")
	if _, err := asSvc.UpdateAutoScalingGroup(blueAsgInput); err != nil {
		log.Printf("updating blue ASG: %v", err)
		setState("failed")
		return
	}

	log.Printf("waiting for blue ASG instances to be ready")
	if err := waitUntilAsgInstancesReady(asSvc, blueAsgName, nAZs); err != nil {
		log.Printf("waiting for blue ASG instances to be ready: %v", err)
		setState("failed")
		return
	}
	log.Printf("blue ASG instances ready")

	// TODO(paulsmith): get this value dynamically from SDK or
	// construct from well-known params
	const targetGroupARN = "arn:aws:elasticloadbalancing:us-east-1:968246069280:targetgroup/paul-example-test/975a4c377b2fa113"

	bluetgi := &autoscaling.AttachLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: aws.String(blueAsgName),
		TargetGroupARNs: []*string{
			aws.String(targetGroupARN),
		},
	}
	log.Printf("putting blue ASG into load")
	if _, err := asSvc.AttachLoadBalancerTargetGroups(bluetgi); err != nil {
		log.Printf("attaching blue ASG to ALB target group: %v", err)
		setState("failed")
		return
	}

	setState("rollforward")

	greentgi := &autoscaling.DetachLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: aws.String(greenAsgName),
		TargetGroupARNs: []*string{
			aws.String(targetGroupARN),
		},
	}
	log.Printf("removing stale green ASG from load")
	if _, err := asSvc.DetachLoadBalancerTargetGroups(greentgi); err != nil {
		log.Printf("detaching green ASG from ALB target group: %v", err)
		setState("failed")
		return
	}

	greenAsgInput := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(greenAsgName),
		DesiredCapacity:      aws.Int64(0),
		MaxSize:              aws.Int64(0),
		MinSize:              aws.Int64(0),
	}
	log.Printf("scaling down stale green ASG")
	if _, err := asSvc.UpdateAutoScalingGroup(greenAsgInput); err != nil {
		log.Printf("updating green ASG: %v", err)
		setState("failed")
		return
	}

	// TODO(paulsmith): there is a race condition because we can't
	// update the tags atomically, so a reader might see both
	// groups as green, or blue, or some indeterminate combination
	// ... risk is pretty low ATM but we should address this
	// somehow later.
	if err := updateTag(asSvc, deployStateTagName, greenAsgName, "blue"); err != nil {
		log.Printf("retagging green ASG: %v", err)
		setState("failed")
		return
	}

	if err := updateTag(asSvc, deployStateTagName, blueAsgName, "green"); err != nil {
		log.Printf("retagging blue ASG: %v", err)
		setState("failed")
		return
	}

	log.Printf("done")

	// TODO(paulsmith): health check

	setState("success")
}

func updateTag(svc *autoscaling.AutoScaling, tagName string, asgName string, value string) error {
	log.Printf("retagging %s to %s:%s", asgName, tagName, value)
	input := &autoscaling.CreateOrUpdateTagsInput{
		Tags: []*autoscaling.Tag{
			{
				Key:               aws.String(tagName),
				ResourceId:        aws.String(asgName),
				ResourceType:      aws.String("auto-scaling-group"),
				Value:             aws.String(value),
				PropagateAtLaunch: aws.Bool(false),
			},
		},
	}
	if _, err := svc.CreateOrUpdateTags(input); err != nil {
		return fmt.Errorf("updating ASG tags: %v", err)
	}
	return nil
}

func waitUntilAsgInstancesReady(svc *autoscaling.AutoScaling, name string, n int) error {
	deadline := time.Now().Add(5 * 60 * time.Second) // fail after 5 minutes
	for {
		count, err := inService(svc, name)
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
