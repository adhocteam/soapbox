provider "aws" {
  region = "${var.region}"
}

resource "aws_s3_bucket" "application_state" {
  bucket = "${var.application_state_bucket}"
  acl    = "private"
}

resource "aws_s3_bucket" "application_releases" {
  bucket = "${var.application_releases_bucket}"
  acl    = "private"
}

resource "aws_dynamodb_table" "application_state_locking" {
  name     = "${var.application_state_table}"
  hash_key = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  write_capacity = 5
  read_capacity  = 5
}
