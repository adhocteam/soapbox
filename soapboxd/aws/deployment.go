package aws

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"text/template"
	"time"

	pb "github.com/adhocteam/soapbox/proto"
	"github.com/adhocteam/soapbox/soapboxd"
	amazon "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

const deployStateTagName = "deploystate"

func (a *aws) deploy(app soapboxd.Application, env soapboxd.Environment, config *pb.Configuration) error {
	// start an ec2 instance, passing a user-data script which
	// installs the docker image and gets the container running

	// Create a userdata script tailored to this application deploy
	var tmpl *template.Template

	root, err := os.Getwd()
	if err != nil {
		return err
	}

	tmpl, err = template.ParseFiles(path.Join(root, "templates", "aws-user-data.tmpl"))
	if err != nil {
		return err
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
		app.Slug,
		// TODO(paulsmith): un-hardcode
		8080,
		soapboxd.SoapboxImageBucket,
		env.Slug,
		fmt.Sprintf("soapbox/%s:%s", app.Slug, app.Committish),
		app.Committish,
		config.ConfigVars,
		strconv.Itoa(int(config.Version)),
	}); err != nil {
		return err
	}

	launchConfig, err := a.createLaunchConfig(app, env, userData.String())
	if err != nil {
		return err
	}

	log.Printf("created launch config: %s", launchConfig)

	var blueASG, greenASG *autoScalingGroup

	blueASG, greenASG, err = a.blueGreenASGs(app, env)
	if err != nil {
		return err
	}

	log.Printf("blue ASG is currently: %s", blueASG.name)
	log.Printf("green ASG is currently: %s", greenASG.name)

	// set number of instances
	minInstances := 2
	maxInstances := 4
	desiredInstances := 2
	for _, v := range config.ConfigVars {
		switch v.Name {
		case "MIN_ASG_INSTANCES":
			if i, err := strconv.Atoi(v.Value); err == nil {
				minInstances = i
			}
		case "MAX_ASG_INSTANCES":
			if i, err := strconv.Atoi(v.Value); err == nil {
				maxInstances = i
			}
		case "DESIRED_ASG_INSTANCES":
			if i, err := strconv.Atoi(v.Value); err == nil {
				desiredInstances = i
			}
		}
	}

	log.Printf("ensuring blue ASG has no instances")

	if err := blueASG.ensureEmpty(); err != nil {
		return err
	}

	log.Printf("updating blue ASG with new launch config")
	if err := blueASG.updateLaunchConfig(launchConfig); err != nil {
		return err
	}

	log.Printf("tagging blue ASG with release info")
	if err := blueASG.updateTags([]tag{{key: "release", value: app.Committish}}); err != nil {
		return err
	}

	log.Printf("starting up blue ASG instances")
	if err := blueASG.resize(minInstances, maxInstances, desiredInstances); err != nil {
		return err
	}

	log.Printf("waiting for blue ASG instances to be ready")
	if err := blueASG.waitUntilInstancesReady(minInstances); err != nil {
		return err
	}

	log.Printf("blue ASG instances ready")
	target, err := greenASG.getTargetGroup(a.sess)
	if err != nil {
		return err
	}

	log.Printf("attaching blue ASG to load balancer")
	if err := blueASG.attachToLBTargetGroup(target.arn); err != nil {
		return err
	}

	log.Printf("waiting for blue ASG instances to pass health checks in load balancer")
	if err := target.waitUntilInstancesReady(blueASG); err != nil {
		return err
	}

	return nil
}

func (a *aws) createLaunchConfig(app soapboxd.Application, env soapboxd.Environment, userData string) (string, error) {
	name := fmt.Sprintf("%s-%s-%s-%d", app.Slug, env.Slug, app.Committish, time.Now().Unix())

	amiID, err := a.getRecentAmiID()
	if err != nil {
		return "", fmt.Errorf("determining ami id: %v", err)
	}

	securityGroupID, err := a.getAppSecurityGroupID(app, env)
	if err != nil {
		return "", err
	}

	input := &autoscaling.CreateLaunchConfigurationInput{
		IamInstanceProfile:      amazon.String(a.config.IamProfile),
		ImageId:                 amazon.String(amiID),
		InstanceType:            amazon.String(a.config.InstanceType),
		KeyName:                 amazon.String(a.config.KeyName),
		LaunchConfigurationName: amazon.String(name),
		SecurityGroups:          []*string{amazon.String(securityGroupID)},
		UserData:                amazon.String(base64.StdEncoding.EncodeToString([]byte(userData))),
	}

	svc := autoscaling.New(a.sess)
	_, err = svc.CreateLaunchConfiguration(input)
	if err != nil {
		return "", fmt.Errorf("creating launch config: %v", err)
	}

	return name, nil
}

func (a *aws) blueGreenASGs(app soapboxd.Application, env soapboxd.Environment) (blue *autoScalingGroup, green *autoScalingGroup, err error) {
	blue, err = a.getASGByColor(app, env, "blue")
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting blue ASG")
	}

	green, err = a.getASGByColor(app, env, "green")
	if err != nil {
		return nil, nil, errors.Wrap(err, "getting green ASG")
	}

	return blue, green, nil
}

func (a *aws) getASGByColor(app soapboxd.Application, env soapboxd.Environment, color string) (*autoScalingGroup, error) {
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
				if *tag.Value == app.Slug {
					found["app"] = true
				}
			case "env":
				if *tag.Value == env.Slug {
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
		return nil, errors.Wrapf(err, "could not find %s ASG in %s environment", color, env.Slug)
	}

	return &autoScalingGroup{
		svc:  svc,
		name: *group.AutoScalingGroupName,
	}, nil
}

func (a *aws) getAppSecurityGroupID(app soapboxd.Application, env soapboxd.Environment) (string, error) {
	sgname := fmt.Sprintf("%s: %s application subnet security group", app.Slug, env.Slug)
	svc := ec2.New(a.sess)
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   amazon.String("group-name"),
				Values: []*string{amazon.String(sgname)},
			},
		},
	}

	res, err := svc.DescribeSecurityGroups(input)
	if err != nil {
		return "", err
	}
	if len(res.SecurityGroups) < 1 {
		return "", errors.New("no security groups found")
	}
	sg := res.SecurityGroups[0]
	return *sg.GroupId, nil
}

func (a *aws) getRecentAmiID() (string, error) {
	svc := ec2.New(a.sess)
	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   amazon.String("virtualization-type"),
			Values: []*string{amazon.String("hvm")},
		},
		&ec2.Filter{
			Name:   amazon.String("name"),
			Values: []*string{amazon.String(a.config.AmiName)},
		},
	}
	descImagesInput := ec2.DescribeImagesInput{
		Filters: filters,
		Owners:  []*string{amazon.String("self")},
	}
	amiRes, err := svc.DescribeImages(&descImagesInput)
	if err != nil {
		fmt.Println(fmt.Sprintf("describing AMIs: %s", err))
		return "", err
	}
	sort.Sort(amiByCreationDate(amiRes.Images))
	return *amiRes.Images[0].ImageId, nil
}

type amiByCreationDate []*ec2.Image

func (a amiByCreationDate) Len() int {
	return len(a)
}

func (a amiByCreationDate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a amiByCreationDate) Less(i, j int) bool {
	return *a[i].CreationDate > *a[j].CreationDate
}

func (a *aws) rollforward(app soapboxd.Application, env soapboxd.Environment) error {
	blueASG, greenASG, err := a.blueGreenASGs(app, env)
	if err != nil {
		return err
	}

	target, err := greenASG.getTargetGroup(a.sess)
	if err != nil {
		return err
	}

	log.Printf("detaching (stale) green ASG from load balancer")
	if err := greenASG.detachFromLBTargetGroup(target.arn); err != nil {
		return err
	}

	// TODO(paulsmith): there is a race condition because we can't
	// update the tags atomically, so a reader might see both
	// groups as green, or blue, or some indeterminate combination
	// ... risk is pretty low ATM but we should address this
	// somehow later.
	log.Printf("swapping blue/green pointers")
	if err := greenASG.updateTags([]tag{{key: deployStateTagName, value: "blue"}}); err != nil {
		return err
	}
	if err := blueASG.updateTags([]tag{{key: deployStateTagName, value: "green"}}); err != nil {
		return err
	}

	return nil
}

func (a *aws) cleanup(app soapboxd.Application, env soapboxd.Environment) {
	log.Printf("cleaning up: terminating instances in blue ASG")
	blueASG, err := a.getASGByColor(app, env, "blue")
	if err != nil {
		log.Printf("getting blue ASG: %v", err)
	}
	if err := blueASG.ensureEmpty(); err != nil {
		log.Printf("ensuring blue ASG is empty: %v", err)
	}
	log.Printf("cleaning up: blue ASG empty")
}
