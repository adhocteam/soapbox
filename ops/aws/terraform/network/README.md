# Soapbox application network configuration

## Usage

1. Plan and apply the network configuration. For example:

```
$ cd ops/terraform/aws/network
$ terraform plan \
  -var 'application_domain=no-name-application.com' \
  -var 'application_name=no-name-application' \
  -var 'environment=test'
```

## Other variables

### instance_tenancy

Default: `default`

Set the [tenancy](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/dedicated-instance.html) for the hardware on which your application runs. Using `default` will probably suit most, however it can be changed to `dedicated` if required.

Example: `terraform apply -var 'instance_tenancy="dedicated"'`

### region

Default: `us-east-1`

Your preferred AWS region.

Example: `terraform apply -var 'region=us-west-1'`

### availability_zones

Default: `["a", "b"]`

Availability zones within the region that should be used.

Example: `terraform apply -var 'availability_zones=["a", "b", "c"]'

NOTE: If you specify more than two availability_zones, you _must_ make corresponding changes to `az_cidr_blocks`.

### public_ingress_cidrs

Default: `0.0.0.0/0`

The origin IPs that are allowed access to your Application Load Balancer (ALB). By default, all requests are allowed.

Example: `terraform apply -var 'public_ingress_cidrs="129.14.129.13/32"'`

### vpc_cidr_block

Default: `10.0.0.0/16`

The adddress space to use for the application's VPC. This default allows for `65536` hosts.

### az_cidr_blocks

Default: `{"app" = ["10.0.0.0/20", "10.0.16.0/20"], "dmz"  = ["10.0.32.0/20", "10.0.48.0/20"]}`

The breakdown of address spaces for each of the zones specified for `availability_zones`.

Using the default values of `region=us-east-1`, `availability_zones = ["a", "b"]` and `app = ["10.0.0.0/20", "10.0.16.0/20"]`, this means that the app subnet in availability zone "us-east-1a" would have allocation "10.0.0.0/20".
