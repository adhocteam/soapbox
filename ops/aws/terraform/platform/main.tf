provider "aws" {
  region = "${var.region}"
}

resource "aws_s3_bucket" "application_releases" {
  bucket = "${var.application_releases_bucket}"
  acl    = "private"
}

resource "aws_s3_bucket" "application_state_bucket" {
  bucket = "${var.application_state_bucket}"
  acl    = "private"
}

resource "aws_dynamodb_table" "application_state_locking_table" {
  name           = "${var.application_state_locking_table}"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "LockId"

  attribute {
    name = "LockId"
    type = "S"
  }
}
