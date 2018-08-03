package aws

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/adhocteam/soapbox"
	amazon "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/pkg/errors"
)

func (a *aws) deploy() error {
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

# Configure logs to forward to rsyslogd
cat << EOF > /var/log/$APP_NAME/config
U127.0.0.1
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
--env PORT \
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
	var tmpl *template.Template
	tmpl, err = template.New("user-data.tmpl").Parse(userDataTmpl)

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
		app.slug,
		// TODO(paulsmith): un-hardcode
		8080,
		soapboxImageBucket,
		env.slug,
		image,
		committish,
		config.ConfigVars,
		strconv.Itoa(int(config.Version)),
	}); err != nil {
		return err
	}

	var securityGroupID string

	securityGroupId, err = app.getAppSecurityGroupId(env)
	if err != nil {
		return err
	}

	var launchConfig string

	launchConfig, err = createLaunchConfig(s.config, app, env, committish, securityGroupId, time.Now(), userData.String())
	if err != nil {
		return err
	}

	log.Printf("created launch config: %s", launchConfig)

	var blueASG, greenASG *autoScalingGroup

	blueASG, greenASG, err = app.blueGreenASGs(env)
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
	do(func() error {
		return blueASG.ensureEmpty()
	})

	log.Printf("updating blue ASG with new launch config")
	do(func() error {
		return blueASG.updateLaunchConfig(launchConfig)
	})

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
	do(func() error {
		return blueASG.updateTags([]tag{{key: "release", value: committish}})
	})

	log.Printf("starting up blue ASG instances")
	do(func() error {
		return blueASG.resize(minInstances, maxInstances, desiredInstances)
	})

	log.Printf("waiting for blue ASG instances to be ready")
	do(func() error { return blueASG.waitUntilInstancesReady(minInstances) })
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
}

func createLaunchConfig(config *soapbox.Config, appSlug string, envSlug string, committish string, securityGroupID string, t time.Time, userData string) (string, error) {
	name := fmt.Sprintf("%s-%s-%s-%d", appSlug, envSlug, committish, t.Unix())

	amiId, err := app.getRecentAmiId(config.AmiName)
	if err != nil {
		return "", fmt.Errorf("determining ami id: %v", err)
	}

	input := &autoscaling.CreateLaunchConfigurationInput{
		IamInstanceProfile:      amazon.String(config.IamProfile),
		ImageId:                 amazon.String(amiId),
		InstanceType:            amazon.String(config.InstanceType),
		KeyName:                 amazon.String(config.KeyName),
		LaunchConfigurationName: amazon.String(name),
		SecurityGroups:          []*string{amazon.String(securityGroupID)},
		UserData:                amazon.String(base64.StdEncoding.EncodeToString([]byte(userData))),
	}

	svc := autoscaling.New(app.sess)
	_, err = svc.CreateLaunchConfiguration(input)
	if err != nil {
		return "", fmt.Errorf("creating launch config: %v", err)
	}

	return name, nil
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
