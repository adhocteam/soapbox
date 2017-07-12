/*
NOTE: Tag values are used to perform lookups of resources
in other terraform configuration files. The variable substitutions
must be kept in sync with corresponding tags for data sources
used in "aws/deployment" configurations.
*/

provider "aws" {
  region = "${var.region}"
}

# VPC
resource "aws_vpc" "application_vpc" {
  cidr_block           = "${var.vpc_cidr_block}"
  instance_tenancy     = "${var.instance_tenancy}"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags {
    Name = "${var.application_name}: ${var.environment} vpc"
    app  = "${var.application_name}"
    env  = "${var.environment}"
  }
}

# Application subnets
resource "aws_subnet" "app" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${aws_vpc.application_vpc.id}"
  cidr_block        = "${element(var.az_cidr_blocks["app"], count.index)}"
  availability_zone = "${var.region}${element(var.availability_zones, count.index)}"

  tags {
    Name = "${var.application_name}: app subnet ${count.index}"
    env  = "${var.environment}"
    app  = "${var.application_name}"
  }
}

# DMZ subnets
resource "aws_subnet" "dmz" {
  count             = "${length(var.availability_zones)}"
  vpc_id            = "${aws_vpc.application_vpc.id}"
  cidr_block        = "${element(var.az_cidr_blocks["dmz"], count.index)}"
  availability_zone = "${var.region}${element(var.availability_zones, count.index)}"

  tags {
    Name = "${var.application_name}: dmz subnet ${count.index}"
    env  = "${var.environment}"
    app  = "${var.application_name}"
  }
}

# Security Group: world -> alb
resource "aws_security_group" "application_alb_sg" {
  name   = "${var.application_name}: ${var.environment} public alb security group"
  vpc_id = "${aws_vpc.application_vpc.id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["${var.public_ingress_cidrs}"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["${var.public_ingress_cidrs}"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.application_name}: ${var.environment} public alb security group"
    env  = "${var.environment}"
    app  = "${var.application_name}"
  }
}

# Security group: alb -> app subnet
resource "aws_security_group" "application_app_sg" {
  name   = "${var.application_name}: ${var.environment} application subnet security group"
  vpc_id = "${aws_vpc.application_vpc.id}"

  ingress {
    from_port       = "${var.application_port}"
    to_port         = "${var.application_port}"
    protocol        = "tcp"
    security_groups = ["${aws_security_group.application_alb_sg.id}"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name = "${var.application_name}: ${var.environment} application subnet security group"
    env  = "${var.environment}"
    app  = "${var.application_name}"
  }
}

# Public alb
resource "aws_alb" "application_alb" {
  name            = "${var.application_name}-${var.environment}"
  internal        = false
  security_groups = ["${aws_security_group.application_alb_sg.id}"]
  subnets         = ["${aws_subnet.dmz.*.id}"]

  ip_address_type = "ipv4"

  tags {
    Name = "${var.application_name}: ${var.environment} alb"
    env  = "${var.environment}"
    app  = "${var.application_name}"
  }
}

# Route53 zone
resource "aws_route53_zone" "application_zone" {
  name    = "${var.application_domain}."
  comment = "${var.application_name}: ${var.environment} domain"
  vpc_id  = "${aws_vpc.application_vpc.id}"

  tags {
    Name = "${var.application_name}: ${var.environment} hosted zone"
    env  = "${var.environment}"
    app  = "${var.application_name}"
  }
}

# IGW
resource "aws_internet_gateway" "application_igw" {
  vpc_id = "${aws_vpc.application_vpc.id}"

  tags {
    Name = "${var.application_name}: ${var.environment} igw"
    env  = "${var.environment}"
    app  = "${var.application_name}"
  }
}

# EIPs for NAT gateways
resource "aws_eip" "application_eip" {
  count = "${length(var.availability_zones)}"
  vpc   = true
}

# NAT gateways for DMZ subnets
resource "aws_nat_gateway" "dmz" {
  count         = "${length(var.availability_zones)}"
  allocation_id = "${element(aws_eip.application_eip.*.id, count.index)}"
  subnet_id     = "${element(aws_subnet.dmz.*.id, count.index)}"

  depends_on = ["aws_internet_gateway.application_igw"]
}
