/*
Use this config as a helper for bootstrapping the
resources required to track Soapbox platform
infrastructure state. These resources should be
created once when configuring the AWS provider
for Soapbox.

Run: terraform apply -state=/dev/null -lock=false
*/

variable "region" {
  type    = "string"
  default = "us-east-1"
}

provider "aws" {
  region = "${var.region}"
}

variable "platform_state_bucket" {
  type    = "string"
  default = "soapbox-platform-tf-state"
}

variable "platform_state_table" {
  type    = "string"
  default = "soapbox-platform-state-locking"
}

resource "aws_s3_bucket" "platform_state" {
  bucket = "${var.platform_state_bucket}"
  acl    = "private"
}

resource "aws_dynamodb_table" "platform_state_locking" {
  name     = "${var.platform_state_table}"
  hash_key = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  write_capacity = 5
  read_capacity  = 5
}
