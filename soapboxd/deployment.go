package soapboxd

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"text/template"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/adhocteam/soapbox"
	"github.com/adhocteam/soapbox/models"
	pb "github.com/adhocteam/soapbox/proto"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/s3"
	gpb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// TODO(paulsmith): platform config
const soapboxImageBucket = "soapbox-app-images"

func (s *server) ListDeployments(ctx context.Context, req *pb.ListDeploymentRequest) (*pb.ListDeploymentResponse, error) {
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
	byId := make(map[int32]*pb.Environment)
	for _, env := range envRes.Environments {
		byId[env.GetId()] = env
	}
	for _, d := range deployments {
		d.Env = byId[d.GetEnv().GetId()]
	}

	res := &pb.ListDeploymentResponse{
		Deployments: deployments,
	}
	return res, nil
}

func (s *server) GetDeployment(ctx context.Context, req *pb.GetDeploymentRequest) (*pb.Deployment, error) {
	return nil, nil
}

// GetLatestDeployment gets latest deployment for an application environment.
func (s *server) GetLatestDeployment(ctx context.Context, req *pb.GetLatestDeploymentRequest) (*pb.Deployment, error) {
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
		} else {
			return nil, errors.Wrap(err, "querying deployments table")
		}
	}
	setPbTimestamp(dep.CreatedAt, createdAt)
	return dep, nil
}

func (s *server) StartDeployment(ctx context.Context, req *pb.Deployment) (*pb.StartDeploymentResponse, error) {
	req.State = "rollout-wait"
	query := `INSERT INTO deployments (application_id, environment_id, committish, current_state) VALUES ($1, $2, $3, $4) RETURNING id`
	appId := int(req.GetApplication().GetId())
	args := []interface{}{
		appId,
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
	app, err := models.ApplicationByID(s.db, appId)
	if err != nil {
		return nil, errors.Wrap(err, "getting application model from db")
	}
	req.Application.Name = app.Name
	req.Application.Description = nullString(app.Description)
	req.Application.GithubRepoUrl = nullString(app.GithubRepoURL)
	req.Application.Slug = app.Slug
	req.Application.Id = int32(app.ID)
	req.Application.UserId = int32(app.UserID)

	envReq := pb.GetEnvironmentRequest{req.GetEnv().GetId()}
	env, err := s.GetEnvironment(ctx, &envReq)
	if err != nil {
		return nil, errors.Wrap(err, "getting environment")
	}
	req.Env = env

	deploy := &deployState{
		id:     int(req.GetId()),
		userID: app.UserID,
	}
	deploy.app = newAppFromProtoBuf(req.GetApplication())
	deploy.sess, err = session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "getting AWS session")
	}
	deploy.env = newEnvFromProtoBuf(req.GetEnv())
	deploy.config, err = s.GetLatestConfiguration(context.TODO(), &pb.GetLatestConfigurationRequest{
		EnvironmentId: deploy.env.id,
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting latest app configuration")
	}
	deploy.committish = req.GetCommittish()

	go s.advanceDeployment(deploy)
	res := &pb.StartDeploymentResponse{
		Id: req.GetId(),
	}

	if err := s.AddDeploymentActivity(context.Background(), pb.ActivityType_DEPLOYMENT_STARTED, deploy); err != nil {
		return nil, err
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

/*

states:

start
rollout-wait
evaluate-wait
rollforward
rollforward-wait
rollback
rollback-wait
success
failure

events:

rollout-started
rollout-in-progress
rollout-ok
evaluate-in-progress
evaluate-ok
rollforward-started
rollforward-in-progress
rollforward-completed
...

*/

type deployState struct {
	id         int
	userID     int
	app        *application
	sess       *session.Session
	env        *environment
	config     *pb.Configuration
	committish string
}

func dockerImageName(app *application, committish string) string {
	return fmt.Sprintf("soapbox/%s:%s", app.slug, committish)
}

func rollout(s *server, dep *deployState, ch chan string) handler {
	return func(state, event) error {
		app := dep.app

		// get a temp dir to work in
		tempdir, err := ioutil.TempDir("", "sandbox")
		if err != nil {
			return errors.Wrap(err, "making temp dir")
		}
		if tempdir != "" {
			defer os.RemoveAll(tempdir)
		}

		// clone the repo at the committish
		const appdir = "appdir"
		cmd := exec.Command("git", "clone", app.githubRepoUrl, appdir)
		cmd.Dir = tempdir
		log.Printf("cloning repo")
		if out, err := cmd.CombinedOutput(); err != nil {
			return errors.Wrapf(err, "running git clone: %s", out)
		}

		committish := dep.committish
		cmd = exec.Command("git", "checkout", committish)
		cmd.Dir = filepath.Join(tempdir, appdir)
		log.Println("checking out committish")
		if out, err := cmd.CombinedOutput(); err != nil {
			return errors.Wrapf(err, "running git checkout: %s", out)
		}

		if sha1Re.MatchString(committish) {
			// use short committish from here on out
			committish = committish[:7]
		}

		image := dockerImageName(app, committish)

		// build the docker image from the repo
		cmd = exec.Command("docker", "build", "-t", image, ".")
		cmd.Dir = filepath.Join(tempdir, appdir)
		log.Printf("building docker image: %s", image)
		if out, err := cmd.CombinedOutput(); err != nil {
			return errors.Wrapf(err, "running docker build: %s", out)
		}

		// export the docker image to a file
		log.Printf("saving docker image %s to file", image)
		filename, err := exportDockerImageToFile(tempdir, image)
		if err != nil {
			return errors.Wrap(err, "exporting docker image to file")
		}

		// upload docker image to S3 bucket
		log.Println("upload docker image to S3")
		objectKey := fmt.Sprintf("%s/%s-%s.tar.gz", app.slug, app.slug, committish)
		storage := newS3Storage(app.sess)
		if err := storage.uploadFile(soapboxImageBucket, objectKey, filename); err != nil {
			return errors.Wrap(err, "uploading docker image to S3")
		}

		ch <- "rollout-finish"
		return nil
	}
}

func evaluate(s *server, deploy *deployState, ch chan string) handler {
	return func(state, event) error {
		app := deploy.app
		env := deploy.env
		committish := deploy.committish
		config := deploy.config

		// start an ec2 instance, passing a user-data script which
		// installs the docker image and gets the container running
		tmpl, err := template.New("user-data.tmpl").Parse(userDataTmpl)
		if err != nil {
			return errors.Wrap(err, "parsing user data template")
		}
		var userData bytes.Buffer
		if err := tmpl.Execute(&userData, struct {
			Slug          string
			ListenPort    int
			Bucket        string
			Environment   string
			Image         string
			Release       string
			Variables     []*pb.ConfigVar
			ConfigVersion string
		}{
			app.slug,
			// TODO(paulsmith): un-hardcode
			8080,
			soapboxImageBucket,
			env.slug,
			dockerImageName(app, committish),
			committish,
			config.ConfigVars,
			strconv.Itoa(int(config.Version)),
		}); err != nil {
			return errors.Wrap(err, "executing user data template")
		}

		securityGroupId, err := app.getAppSecurityGroupId(env)
		if err != nil {
			return errors.Wrap(err, "getting app security group ID")
		}

		launchConfig, err := createLaunchConfig(
			s.config,
			app,
			env,
			committish,
			securityGroupId,
			time.Now(),
			userData.String(),
		)
		if err != nil {
			return errors.Wrap(err, "creating launch config")
		}

		log.Printf("created launch config: %s", launchConfig)

		blueASG, greenASG, err := app.blueGreenASGs(env)
		if err != nil {
			return errors.Wrap(err, "creating blue & green ASGs")
		}

		log.Printf("blue ASG is currently: %s", blueASG.name)
		log.Printf("green ASG is currently: %s", greenASG.name)

		const nAZs = 2 // number of availability zones

		log.Printf("ensuring blue ASG has no instances")
		if err := blueASG.ensureEmpty(); err != nil {
			return errors.Wrap(err, "ensuring empty blue ASG")
		}

		log.Printf("updating blue ASG with new launch config")
		if err := blueASG.updateLaunchConfig(launchConfig); err != nil {
			return errors.Wrap(err, "updating blue ASG launch config")
		}

		defer func() {
			log.Printf("cleaning up: terminating instances in blue ASG")
			blueASG, err := app.getASGByColor(env, "blue")
			if err != nil {
				log.Printf("getting blue ASG: %v", err)
			}
			if err := blueASG.ensureEmpty(); err != nil {
				log.Printf("ensuring blue ASG is empty: %v", err)
			}
			log.Printf("cleaning up: blue ASG empty")
		}()

		log.Printf("tagging blue ASG with release info")
		if err := blueASG.updateTags([]tag{{key: "release", value: committish}}); err != nil {
			return errors.Wrap(err, "updating blue ASG tags")
		}

		log.Printf("starting up blue ASG instances")
		if err := blueASG.resize(nAZs, nAZs*2, nAZs); err != nil {
			return errors.Wrap(err, "resizing blue ASG")
		}

		log.Printf("waiting for blue ASG instances to be ready")
		if err := blueASG.waitUntilInstancesReady(nAZs); err != nil {
			return errors.Wrap(err, "waiting until blue ASG instances ready")
		}
		log.Printf("blue ASG instances ready")

		target, err := greenASG.getTargetGroup()
		if err != nil {
			return errors.Wrap(err, "getting target group")
		}

		log.Printf("attaching blue ASG to load balancer")
		if err := blueASG.attachToLBTargetGroup(target.arn); err != nil {
			return errors.Wrap(err, "attaching blue ASG to ALB target group")
		}

		log.Printf("waiting for blue ASG instances to pass health checks in load balancer")
		if err := target.waitUntilInstancesReady(blueASG); err != nil {
			return errors.Wrap(err, "waiting until instances ready in ALB target")
		}

		ch <- "evaluate-finish"
		return nil
	}
}

func rollforward(s *server, deploy *deployState, ch chan string) handler {
	return func(state, event) error {
		app := deploy.app
		env := deploy.env

		blueASG, greenASG, err := app.blueGreenASGs(env)
		if err != nil {
			return errors.Wrap(err, "creating blue & green ASGs")
		}

		target, err := greenASG.getTargetGroup()
		if err != nil {
			return errors.Wrap(err, "getting target group")
		}

		log.Printf("detaching (stale) green ASG from load balancer")
		if err := greenASG.detachFromLBTargetGroup(target.arn); err != nil {
			return errors.Wrap(err, "detaching green ASG from ALB target")
		}

		// TODO(paulsmith): there is a race condition because we can't
		// update the tags atomically, so a reader might see both
		// groups as green, or blue, or some indeterminate combination
		// ... risk is pretty low ATM but we should address this
		// somehow later.
		log.Printf("swapping blue/green pointers")
		if err := greenASG.updateTags([]tag{{key: deployStateTagName, value: "blue"}}); err != nil {
			return errors.Wrap(err, "updating tags on green ASG")
		}
		if err := blueASG.updateTags([]tag{{key: deployStateTagName, value: "green"}}); err != nil {
			return errors.Wrap(err, "updating tags on blue ASG")
		}

		log.Printf("done")

		// TODO(paulsmith): health check?

		ch <- "rollforward-finish"
		return nil
	}
}

func (s *server) advanceDeployment(deploy *deployState) {
	events := make(chan string)
	m := newFsm("start")
	m.addTransition("start", "rollout-start", "rollout-wait", rollout(s, deploy, events))
	m.addTransition("rollout-wait", "rollout-finish", "evaluate-wait", evaluate(s, deploy, events))
	m.addTransition("evaluate-wait", "evaluate-finish", "rollforward-wait", rollforward(s, deploy, events))

	go func() {
		if m.state == "start" {
			events <- "rollout-start"
		}
	}()

	for {
		select {
		case ev := <-events:
			if err := m.step(event(ev)); err != nil {
				if err := s.AddDeploymentActivity(context.Background(), pb.ActivityType_DEPLOYMENT_FAILURE, deploy); err != nil {
					log.Printf("adding deployment activity: %v", err)
				}
				// failed. ....
			}
		}
	}

	if err := s.AddDeploymentActivity(context.Background(), pb.ActivityType_DEPLOYMENT_SUCCESS, deploy); err != nil {
		log.Printf("error adding deployment activity", err)
	}
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
	deadline := time.Now().Add(10 * time.Minute)
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

func (g *autoScalingGroup) updateLaunchConfig(lcName string) error {
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName:    aws.String(g.name),
		LaunchConfigurationName: aws.String(lcName),
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

func (g *autoScalingGroup) ensureEmpty() error {
	insts, err := g.getInstances()
	if err != nil {
		return errors.Wrap(err, "getting instances")
	}
	if len(insts) == 0 {
		return nil
	}
	if err := g.resize(0, 0, 0); err != nil {
		return errors.Wrap(err, "resizing group to 0")
	}
	if err := g.waitUntilGroupEmpty(); err != nil {
		return errors.Wrap(err, "waiting until group empty")
	}
	return nil
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

func createLaunchConfig(config *soapbox.Config, app *application, env *environment, committish string, securityGroupId string, t time.Time, userData string) (string, error) {
	name := fmt.Sprintf("%s-%s-%s-%d", app.slug, env.slug, committish, t.Unix())

	amiId, err := app.getRecentAmiId(config.AmiName)
	if err != nil {
		return "", fmt.Errorf("determining ami id: %v", err)
	}

	input := &autoscaling.CreateLaunchConfigurationInput{
		IamInstanceProfile:      aws.String(config.IamProfile),
		ImageId:                 aws.String(amiId),
		InstanceType:            aws.String(config.InstanceType),
		KeyName:                 aws.String(config.KeyName),
		LaunchConfigurationName: aws.String(name),
		SecurityGroups:          []*string{aws.String(securityGroupId)},
		UserData:                aws.String(base64.StdEncoding.EncodeToString([]byte(userData))),
	}

	svc := autoscaling.New(app.sess)
	_, err = svc.CreateLaunchConfiguration(input)
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
	id            int
	name          string
	slug          string
	githubRepoUrl string

	sess *session.Session
}

func newAppFromProtoBuf(appPb *pb.Application) *application {
	return &application{
		id:            int(appPb.GetId()),
		name:          appPb.GetName(),
		slug:          appPb.GetSlug(),
		githubRepoUrl: appPb.GetGithubRepoUrl(),
	}
}

type environment struct {
	id   int32
	name string
	slug string
}

func newEnvFromProtoBuf(envPb *pb.Environment) *environment {
	return &environment{
		id:   envPb.GetId(),
		name: envPb.GetName(),
		slug: envPb.GetSlug(),
	}
}

func (a *application) getASGByColor(env *environment, color string) (*autoScalingGroup, error) {
	// get a list of all ASGs and iterate over until find
	// "deploystate" tags for our app and environment

	svc := autoscaling.New(a.sess)
	asgs, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return nil, errors.Wrap(err, "describing ASGs")
	}

	var group *autoscaling.Group

	for _, asg := range asgs.AutoScalingGroups {
		found := make(map[string]bool)
		for _, tag := range asg.Tags {
			switch *tag.Key {
			case "app":
				if *tag.Value == a.slug {
					found["app"] = true
				}
			case "env":
				if *tag.Value == env.slug {
					found["env"] = true
				}
			case deployStateTagName:
				if *tag.Value == color {
					found["deploystate"] = true
				}
			}
		}
		if found["app"] && found["env"] && found["deploystate"] {
			group = asg
			break
		}
	}

	if group == nil {
		return nil, errors.Wrapf(err, "could not find %s ASG in %s environment", color, env.slug)
	}

	return &autoScalingGroup{
		sess: a.sess,
		svc:  svc,
		name: *group.AutoScalingGroupName,
	}, nil
}

func (a *application) blueGreenASGs(env *environment) (blue *autoScalingGroup, green *autoScalingGroup, err error) {
	blue, err = a.getASGByColor(env, "blue")
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting blue ASG")
	}

	green, err = a.getASGByColor(env, "green")
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting green ASG")
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

func (a *application) getRecentAmiId(amiNameGlob string) (string, error) {
	svc := ec2.New(a.sess)
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("virtualization-type"),
			Values: []*string{aws.String("hvm")},
		},
		&ec2.Filter{
			Name:   aws.String("name"),
			Values: []*string{aws.String(amiNameGlob)},
		},
	}
	descImagesInput := ec2.DescribeImagesInput{
		Filters: filters,
		Owners:  []*string{aws.String("self")},
	}
	amiRes, err := svc.DescribeImages(&descImagesInput)
	if err != nil {
		fmt.Println(fmt.Sprintf("describing AMIs: %s", err))
		return "", err
	}
	sort.Sort(AmiByCreationDate(amiRes.Images))
	return *amiRes.Images[0].ImageId, nil
}

type AmiByCreationDate []*ec2.Image

func (a AmiByCreationDate) Len() int {
	return len(a)
}

func (a AmiByCreationDate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a AmiByCreationDate) Less(i, j int) bool {
	return *a[i].CreationDate > *a[j].CreationDate
}

// wait for instances to be marked in-service in the ASG lifecycle
func (g *autoScalingGroup) waitUntilInstancesReady(n int) error {
	deadline := time.Now().Add(10 * time.Minute)
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
	return lifecycleState("InService", svc, name)
}

func lifecycleState(state string, svc *autoscaling.AutoScaling, name string) (int, error) {
	out, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(name)},
	})
	if err != nil {
		return 0, err
	}

	count := 0
	group := out.AutoScalingGroups[0]
	for _, inst := range group.Instances {
		if *inst.LifecycleState == state {
			count++
		}
	}
	return count, nil
}

func (g *autoScalingGroup) waitUntilGroupEmpty() error {
	deadline := time.Now().Add(10 * time.Minute)
	for {
		instances, err := g.getInstances()
		if err != nil {
			return errors.Wrap(err, "getting group's instances")
		}
		if len(instances) == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for ASG to be empty)")
		}
		time.Sleep(5 * time.Second)
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

var userDataTmpl = `#!/bin/bash

set -xeuo pipefail

# log all script output
exec > >(tee /var/log/user-data.log) 2>&1

AWS=/usr/bin/aws
DOCKER=/usr/bin/docker

APP_NAME="{{.Slug}}"
PORT="{{.ListenPort}}"
RELEASE_BUCKET="{{.Bucket}}"
RELEASE="{{.Release}}" # Version string/committish
ENV="{{.Environment}}"
IMAGE="{{.Image}}"
CONFIG_VERSION="{{.ConfigVersion}}"

# Retrieve the release from s3
$AWS s3 cp s3://$RELEASE_BUCKET/$APP_NAME/$APP_NAME-$RELEASE.tar.gz /tmp/$APP_NAME-$RELEASE.tar.gz

# Install the docker image
$DOCKER image load -i /tmp/$APP_NAME-$RELEASE.tar.gz

# Set up the runit dirs
mkdir -p "/etc/sv/$APP_NAME"
mkdir -p "/etc/sv/$APP_NAME/env"

# Place env vars in /etc/sv/$APP_NAME/env
{{- range .Variables}}
cat << EOF > /etc/sv/$APP_NAME/env/{{.Name}}
{{.Value}}
EOF
{{end}}

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
exec 2>&1 chpst -e /etc/sv/$APP_NAME/env $DOCKER run \
{{range .Variables -}}
	--env {{.Name}} \
{{end -}}
--rm --name $APP_NAME-run -p 9090:$PORT "$IMAGE"
EOF

# Mark the run file executable
chmod +x /etc/sv/$APP_NAME/run

# Create a link from /etc/service/$APP_NAME -> /etc/sv/$APP_NAME
ln -s /etc/sv/$APP_NAME /etc/service/$APP_NAME

# Switch to /etc/nginx/app.conf
mv /etc/nginx/app.conf /etc/nginx/nginx.conf

# nginx is now proxying to the app itself
service nginx reload

# Set the X-Soapbox-App-Version HTTP header
sed -i.bak \
  $"s/add_header X-Soapbox-App-Version \"latest\"/add_header X-Soapbox-App-Version \"$RELEASE\";\nadd_header X-Soapbox-Config-Version \"$CONFIG_VERSION\";\nadd_header X-Soapbox-Environment \"$ENV\"/" \
  /etc/nginx/nginx.conf

# Safely remove backup
rm -f /etc/nginx/nginx.conf.bak

# Pick up changes to response header
service nginx reload
`
