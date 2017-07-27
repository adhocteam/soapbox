/*
NOTE: Tags are used to perform lookups of data sources
in this configuration file. The variable substitutions
must be kept in sync with corresponding tags for resources
defined in "aws/network" configurations.
*/

data "aws_ami" "soapbox_ami" {
  most_recent = true

  filter {
    name   = "name"
    values = ["soapbox-aws-linux-ami-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["self"]
}

data "aws_vpc" "application_vpc" {
  tags {
    Name = "${var.application_name}: ${var.environment} vpc"
    app  = "${var.application_name}"
    env  = "${var.environment}"
  }
}

data "aws_security_group" "application_app_sg" {
  vpc_id = "${data.aws_vpc.application_vpc.id}"
  name   = "${var.application_name}: ${var.environment} application subnet security group"
}

resource "aws_launch_configuration" "launch_config" {
  name                        = "${var.application_name}-${var.environment}-${var.release}"
  instance_type               = "${var.instance_type}"
  image_id                    = "${data.aws_ami.soapbox_ami.id}"
  security_groups             = ["${data.aws_security_group.application_app_sg.id}"]
  associate_public_ip_address = false
  ebs_optimized               = false
  key_name                    = "${var.key_name}"
  iam_instance_profile        = "${var.instance_profile}"

  lifecycle {
    create_before_destroy = true
  }
}
