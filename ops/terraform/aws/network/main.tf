provider "aws" {
  region = "${var.default_region}"
}

# VPC
resource "aws_vpc" "application_vpc" {
  cidr_block       = "${var.vpc_cidr_block}"
  instance_tenancy = "dedicated"

  tags {
    Name = "${var.application_name}: ${var.environment} vpc"
    env  = "${var.environment}"
  }
}

# Subnets

# Application subnets
resource "aws_subnet" "app_a" {
  vpc_id            = "${aws_vpc.application_vpc.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "${var.default_region}a"

  tags {
    Name = "${var.application_name}: app"
    env  = "${var.environment}"
  }
}

resource "aws_subnet" "app_b" {
  vpc_id     = "${aws_vpc.application_vpc.id}"
  cidr_block = "10.0.2.0/24"
  availability_zone = "${var.default_region}b"

  tags {
    Name = "${var.application_name}: app"
    env  = "${var.environment}"
  }
}

# Data subnets
resource "aws_subnet" "data_a" {
  vpc_id     = "${aws_vpc.application_vpc.id}"
  cidr_block = "10.0.3.0/24"
  availability_zone = "${var.default_region}a"

  tags {
    Name = "${var.application_name}: data"
    env  = "${var.environment}"
  }
}

resource "aws_subnet" "data_b" {
  vpc_id     = "${aws_vpc.application_vpc.id}"
  cidr_block = "10.0.4.0/24"
  availability_zone = "${var.default_region}b"

  tags {
    Name = "${var.application_name}: data"
    env  = "${var.environment}"
  }
}

# DMS subnets
resource "aws_subnet" "dmz_a" {
  vpc_id     = "${aws_vpc.application_vpc.id}"
  cidr_block = "10.0.5.0/24"
  availability_zone = "${var.default_region}a"

  tags {
    Name = "${var.application_name}: dmz"
    env  = "${var.environment}"
  }
}

resource "aws_subnet" "dmz_b" {
  vpc_id     = "${aws_vpc.application_vpc.id}"
  cidr_block = "10.0.6.0/24"
  availability_zone = "${var.default_region}b"

  tags {
    Name = "${var.application_name}: ${var.environment} dmz"
    env  = "${var.environment}"
  }
}

# Security Group: world -> alb
resource "aws_security_group" "application_alb_sg" {
  name = "${var.application_name}: ${var.environment} public alb security group"

  ingress {
    from_port   = "${var.application_port}"
    to_port     = "${var.application_port}"
    protocol    = "tcp"
    cidr_blocks = ["${var.public_ingress_cidrs}"]
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    cidr_blocks     = ["0.0.0.0/0"]
  }
}

# Security group: alb -> app subnet
resource "aws_security_group" "application_app_sg" {
  name = "${var.application_name}: ${var.environment} application subnet security group"

  ingress {
    from_port       = "${var.application_port}"
    to_port         = "${var.application_port}"
    protocol        = "tcp"
    security_groups = ["${aws_security_group.application_alb_sg.id}"]
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    cidr_blocks     = ["0.0.0.0/0"]
  }
}

# Security group: app subnet <-> data subnet
resource "aws_security_group" "application_data_sg" {
  name = "${var.application_name}: ${var.environment} data subnet security group"

  ingress {
    from_port       = "${var.application_data_port}"
    to_port         = "${var.application_data_port}"
    protocol        = "tcp"
    security_groups = ["${aws_security_group.application_app_sg.id}"]
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "tcp"
    security_groups = ["${aws_security_group.application_app_sg.id}"]
  }
}

# Public alb
resource "aws_alb" "application_alb" {
  name = "application-alb-${var.environment}"
  internal = false
  security_groups = ["${aws_security_group.application_alb_sg.id}"]
  subnets = ["${aws_subnet.dmz_a.id}", "${aws_subnet.dmz_b.id]"]

  ip_address_type = "ipv4"

  tags {
    Name = "${var.application_name}: ${var.environment} alb"
    env  = "${var.environment}"
  }
}

# Route53 zone
resource "aws_route53_zone" "application_zone" {
  name = "${var.application_domain}."
  comment = "${var.application_name}: ${var.env} domain"

  tags {
    Name = "${var.application_name}: ${var.environment} hosted zone"
    env  = "${var.environment}"
  }
}

# IGW
resource "aws_internet_gateway" "application_igw" {
  vpc_id = "${aws_vpc.application_vpc.id}"

  tags {
    Name = "${var.application_name}: ${var.environment} igw"
    env  = "${var.environment}"
  }
}

# EIPs for NAT gateways
resource "aws_eip" "application_eip_one" {
  vpc = true

  tags {
    Name = "${var.application_name}: ${var.environment} EIP 1"
    env  = "${var.environment}"
  }
}

resource "aws_eip" "application_eip_two" {
  vpc = true

  tags {
    Name = "${var.application_name}: ${var.environment} EIP 2"
    env  = "${var.environment}"
  }
}

# NAT gateways for DMZ subnets
resource "aws_nat_gateway" "dmz_a" {
  allocation_id = "${aws_eip.application_eip_one.id}"
  subnet_id     = "${aws_subnet.dmz_a.id}"

  depends_on = ["aws_internet_gateway.application_igw"]

  tags {
    Name = "${var.application_name}: ${var.environment} nat gw ${var.default_region}a"
    env  = "${var.environment}"
  }
}

resource "aws_nat_gateway" "dmz_b" {
  allocation_id = "${aws_eip.application_eip_two.id}"
  subnet_id     = "${aws_subnet.dmz_b.id}"

  depends_on = ["aws_internet_gateway.application_igw"]

  tags {
    Name = "${var.application_name}: ${var.environment} nat gw ${var.default_region}b"
    env  = "${var.environment}"
  }
}
