provider "aws" {
  region = "${var.region}"
}

resource "aws_s3_bucket" "application_releases" {
  bucket = "${var.application_releases_bucket}"
  acl    = "private"
}

resource "aws_s3_bucket" "application_configs" {
  bucket = "${var.application_configs_bucket}"
  acl    = "private"
}

resource "aws_s3_bucket" "application_state_bucket" {
  bucket = "${var.application_state_bucket}"
  acl    = "private"
}

resource "aws_dynamodb_table" "application_state_table" {
  name           = "${var.application_state_table}"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

resource "aws_route53_zone" "platform_zone" {
  name    = "${var.platform_domain}."
  comment = "Soapbox platform hosted zone: ${var.platform_domain}"

  tags {
    Name = "Soapbox platform hosted zone: ${var.platform_domain}"
  }
}
