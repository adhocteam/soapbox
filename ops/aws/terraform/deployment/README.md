# Soapbox application deployment config

## Usage

After creating an application's network infrastructure, the next step is to
create its auto scaling groups, initial launch configuration, ALB target group, etc.

```
$ cd ops/aws/terraform/deployment/asg
$ terraform get
$ terraform apply \
  -var "application_name=no-name-application" \
  -var "environment=test" \
  -var "release=1"
```

## Other variables

NOTE: where variable names match those of the application network config files,
their values should also match.

### region

Default: `us-east-1`

Your preferred AWS region.

Example: `terraform apply -var 'region=us-west-1'`

### availability_zones

Default: `["a", "b"]`

Availability zones within the region that should be used.

Example: `terraform apply -var 'availability_zones=["a", "b", "c"]'

### green_asg_max

Default: `6`

The maximum number of instances the "green" autoscaling group can use.

Example: `terraform apply -var 'green_asg_max=12'

### green_asg_min

Default: `1`

The minimum number of instances the "green" autoscaling group can use.

Example: `terraform apply -var 'green_asg_min=2'

### green_asg_desired

Default: `1`

The ideal/desired number of instances in the "green" autoscaling group.

Example: `terraform apply -var 'green_asg_desired=2'

### application_acm_cert_arn

Default: `(empty string)`

The ARN of an Amazon Certificate Manager certificate used to enable `https` listener on your application's ALB.

Example: `terraform apply -var 'application_acm_cert_arn=${ARN_GOES_HERE}'

### application_port

Default: `8080`

The docker port that your application uses.

Example: `terraform apply -var 'application_port=3000'

### health_check_path

Default: `/_health`

The endpoint used to check that your application is functioning properly. Should return an `200` status code if all is well.

Example: `terraform apply -var 'health_check_path=/_healthcheck'
