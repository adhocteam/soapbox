# Soapbox platform and Terraform

## Getting started

Soapbox includes a handful of Terraform configuration files to
ease the process of bootstrapping and tracking changes to platform
infrastructure.

Initial set up of a Soapbox (platform) deployment requires
creation of some AWS resources used to track and manage
the state of _other_ AWS (or other cloud provider) resources.

The config provided in `aws/terraform/platform/boostrap` can
be used for this process.

In the commands that follow, modify `my-unique-platform-state-s3-bucket`
to use a unique name for the S3 bucket that will be used to
store Soapbox's terraform state.

```
$ cd aws/terraform/platform/boostrap
# To view the resources to be created
$ terraform plan \
    -var "platform_state_bucket=my-unique-platform-state-s3-bucket"
    -state=/dev/null \
    -lock=false
# To actually create them
$ terraform apply \
    -var "platform_state_bucket=my-unique-platform-state-s3-bucket"
    -state=/dev/null \
    -lock=false
```

## Creating and tracking Soapbox platform infrastructure

Once the requisite AWS resources have been created, you
can create and track changes to Soapbox platform infrastructure.

1. Copy `backend.tfvars.sample` to `backend.tfvars`, modifying the `bucket`
value to match the name you used in place of `my-unique-platform-state-s3-bucket`
above.

2. Initialize terraform:

```
$ cd aws/terraform/platform
$ terraform init -backend-config=backend.tfvars
```

3. Plan and apply to create Soapbox resources. Modify `my-unique-app-images-s3-bucket`
in the commands below to use a unique name for the S3 bucket that will
be used to store Soapbox application releases.

```
# To view the resources to be created
$ terraform plan -var "application_releases_bucket=my-unique-app-images-s3-bucket"
# To actually create them
$ terraform apply -var "application_releases_bucket=my-unique-app-images-s3-bucket"
```
