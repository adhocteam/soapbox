/*
NOTE: Tag values are used to perform lookups of resources
in other terraform configuration files. The variable substitutions
must be kept in sync with corresponding tags for data sources
used in "aws/deployment" configurations.
*/

provider "aws" {
  region = "${var.region}"
}

# KMS Application Key (for encrypting and decrypting configurations / secrets)
resource "aws_kms_key" "config_encryption_key" {
  description = "Encryption key used to access application configurations."
  tags {
    Name = "${var.application_name}: ${var.environment} kms encryption key"
    app  = "${var.application_name}"
    env  = "${var.environment}"
  }
}

resource "aws_kms_alias" "config_encryption_key" {
  name          = "alias/${var.application_name}-${var.environment}"
  target_key_id = "${aws_kms_key.config_encryption_key.key_id}"
}

output "aws_kms_arn" {
  sensitive = true
  value     = "${aws_kms_key.config_encryption_key.arn}"
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
    from_port       = 80
    to_port         = 80
    protocol        = "tcp"
    security_groups = ["${aws_security_group.application_alb_sg.id}"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${aws_subnet.dmz.*.cidr_block}"]
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

# Lookup of Route53 platform hosted zone
data "aws_route53_zone" "platform_zone" {
  name = "${var.platform_domain}."
}

# Route53 record for <app-name>.<env>.soapbox.hosting
resource "aws_route53_record" "application_subdomain" {
  zone_id = "${data.aws_route53_zone.platform_zone.zone_id}"
  name    = "${var.application_name}.${var.environment}"
  type    = "CNAME"
  ttl     = "300"

  records = ["${aws_alb.application_alb.dns_name}"]
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

resource "aws_route_table" "dmz_subnet_route_table" {
  vpc_id = "${aws_vpc.application_vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.application_igw.id}"
  }

  tags {
    Name = "DMZ subnet route table"
  }
}

resource "aws_route_table_association" "dmz_route_table_assoc" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.dmz.*.id, count.index)}"
  route_table_id = "${aws_route_table.dmz_subnet_route_table.id}"
}

# NAT gateways for DMZ subnets
resource "aws_nat_gateway" "dmz" {
  count         = "${length(var.availability_zones)}"
  allocation_id = "${element(aws_eip.application_eip.*.id, count.index)}"
  subnet_id     = "${element(aws_subnet.dmz.*.id, count.index)}"

  depends_on = ["aws_internet_gateway.application_igw"]
}

resource "aws_route_table" "app_subnet_route_table" {
  count  = "${length(var.availability_zones)}"
  vpc_id = "${aws_vpc.application_vpc.id}"

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = "${element(aws_nat_gateway.dmz.*.id, count.index)}"
  }

  tags = {
    Name = "App subnet route table: ${format("%s%s", var.region, element(var.availability_zones, count.index))}"
  }
}

resource "aws_route_table_association" "app_route_table_assoc" {
  count          = "${length(var.availability_zones)}"
  subnet_id      = "${element(aws_subnet.app.*.id, count.index)}"
  route_table_id = "${element(aws_route_table.app_subnet_route_table.*.id, count.index)}"
}
