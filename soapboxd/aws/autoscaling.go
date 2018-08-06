package aws

import (
	"fmt"
	"time"

	amazon "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/pkg/errors"
)

type autoScalingGroup struct {
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
		AutoScalingGroupName:    amazon.String(g.name),
		LaunchConfigurationName: amazon.String(lcName),
	}
	_, err := g.svc.UpdateAutoScalingGroup(input)
	return err
}

func (g *autoScalingGroup) attachToLBTargetGroup(targetGroupARN string) error {
	input := &autoscaling.AttachLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: amazon.String(g.name),
		TargetGroupARNs: []*string{
			amazon.String(targetGroupARN),
		},
	}
	_, err := g.svc.AttachLoadBalancerTargetGroups(input)
	return err
}

func (g *autoScalingGroup) getTargetGroup(sess *session.Session) (*targetGroup, error) {
	input := &autoscaling.DescribeLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: amazon.String(g.name),
	}
	res, err := g.svc.DescribeLoadBalancerTargetGroups(input)
	if err != nil {
		return nil, err
	}
	group := res.LoadBalancerTargetGroups[0]
	target := &targetGroup{
		svc: elbv2.New(sess),
		arn: *group.LoadBalancerTargetGroupARN,
	}
	return target, nil
}

func (g *autoScalingGroup) detachFromLBTargetGroup(targetGroupARN string) error {
	input := &autoscaling.DetachLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: amazon.String(g.name),
		TargetGroupARNs: []*string{
			amazon.String(targetGroupARN),
		},
	}
	_, err := g.svc.DetachLoadBalancerTargetGroups(input)
	return err
}

func (g *autoScalingGroup) ensureEmpty() error { // Application is a minimal version of the information about an app on the platform
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
		AutoScalingGroupName: amazon.String(g.name),
		DesiredCapacity:      amazon.Int64(int64(desired)),
		MaxSize:              amazon.Int64(int64(max)),
		MinSize:              amazon.Int64(int64(min)),
	}
	_, err := g.svc.UpdateAutoScalingGroup(input)
	return err
}

func (g *autoScalingGroup) getInstances() ([]*autoscaling.Instance, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{amazon.String(g.name)},
	}
	res, err := g.svc.DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, fmt.Errorf("describing ASG: %v", err)
	}
	group := res.AutoScalingGroups[0]
	return group.Instances, nil
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

func inService(svc *autoscaling.AutoScaling, name string) (int, error) {
	return lifecycleState("InService", svc, name)
}

func lifecycleState(state string, svc *autoscaling.AutoScaling, name string) (int, error) {
	out, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{amazon.String(name)},
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
		TargetGroupArn: amazon.String(g.arn),
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
		Key:               amazon.String(t.key),
		ResourceId:        amazon.String(name),
		ResourceType:      amazon.String("auto-scaling-group"),
		Value:             amazon.String(t.value),
		PropagateAtLaunch: amazon.Bool(t.propagateAtLaunch),
	}
}
