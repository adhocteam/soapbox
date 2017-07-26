provider "aws" {
  region = "${var.region}"
}

resource "aws_s3_bucket" "application_releases" {
  bucket = "${var.application_releases_bucket}"
  acl    = "private"
}

resource "aws_route53_zone" "platform_zone" {
  name    = "${var.platform_domain}."
  comment = "Soapbox platform hosted zone: ${var.platform_domain}"

  tags {
    Name = "Soapbox platform hosted zone: ${var.platform_domain}"
  }
}
