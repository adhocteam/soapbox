provider "aws" {
  region = "${var.region}"
}

resource "aws_s3_bucket" "application_releases" {
  bucket = "${var.application_releases_bucket}"
  acl    = "private"
}
