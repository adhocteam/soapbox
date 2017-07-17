/*
NOTE: Tags are used to perform lookups of data sources
in this configuration file. The variable substitutions
must be kept in sync with corresponding tags for resources
defined in "aws/network" configurations.
*/

provider "aws" {
  region = "${var.region}"
}

# Application VPC, subnet and ALB lookup
data "aws_vpc" "application_vpc" {
  tags {
    Name = "${var.application_name}: ${var.environment} vpc"
    app  = "${var.application_name}"
    env  = "${var.environment}"
  }
}

data "aws_subnet" "application_subnet" {
  count  = "${length(var.availability_zones)}"
  vpc_id = "${data.aws_vpc.application_vpc.id}"

  tags {
    Name = "${var.application_name}: app subnet ${count.index}"
    env  = "${var.environment}"
    app  = "${var.application_name}"
  }
}

data "aws_alb" "application_alb" {
  name = "${var.application_name}-${var.environment}"
}

# Launch configuration
module "launch_config" {
    source = "../lc"
    application_name = "${var.application_name}"
    application_port = "${var.application_port}"
    environment      = "${var.environment}"
    release          = "${var.release}"
}

# ALB target group and listeners
resource "aws_alb_target_group" "application_target_group" {
  name     = "${var.application_name}-${var.environment}"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${data.aws_vpc.application_vpc.id}"

  health_check {
    interval            = 60
    path                = "${var.health_check_path}"
    port                = 80
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }

  tags {
    Name = "${var.application_name}: ${var.environment} alb target group"
    env  = "${var.environment}"
    app  = "${var.application_name}"
  }
}

resource "aws_alb_listener" "application_alb_http" {
  load_balancer_arn = "${data.aws_alb.application_alb.arn}"
  port              = 80
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_alb_target_group.application_target_group.arn}"
    type             = "forward"
  }
}

resource "aws_alb_listener" "application_alb_https" {
  count             = "${var.application_acm_cert_arn != "" ? 1 : 0}"
  load_balancer_arn = "${data.aws_alb.application_alb.arn}"
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "${var.application_acm_cert_arn}"

  default_action {
    target_group_arn = "${aws_alb_target_group.application_target_group.arn}"
    type             = "forward"
  }
}

# Blue-green autoscaling groups
resource "aws_autoscaling_group" "asg_blue" {
  availability_zones        = ["${formatlist("%s%s", var.region, var.availability_zones)}"]
  name                      = "${var.application_name}-${var.environment}-blue"
  max_size                  = 0
  min_size                  = 0
  desired_capacity          = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  launch_configuration      = "${module.launch_config.name}"
  target_group_arns         = ["${aws_alb_target_group.application_target_group.arn}"]
  vpc_zone_identifier       = ["${data.aws_subnet.application_subnet.*.id}"]
  enabled_metrics           = [
    "GroupMinSize", "GroupMaxSize", "GroupDesiredCapacity",
    "GroupInServiceInstances", "GroupPendingInstances", "GroupStandbyInstances",
    "GroupTerminatingInstances", "GroupTotalInstances"
  ]

  tags = [
    {
      key                 = "Name"
      value               = "${var.application_name}-${var.environment}"
      propagate_at_launch = true
    },
    {
      key                 = "app"
      value               = "${var.application_name}"
      propagate_at_launch = true
    },
    {
      key                 = "env"
      value               = "${var.environment}"
      propagate_at_launch = true
    },
    {
      key                 = "release"
      value               = "latest"
      propagate_at_launch = true
    }
  ]

  lifecycle {
    ignore_changes = ["launch_configuration"]
  }
}

resource "aws_autoscaling_group" "asg_green" {
  availability_zones        = ["${formatlist("%s%s", var.region, var.availability_zones)}"]
  name                      = "${var.application_name}-${var.environment}-green"
  max_size                  = "${var.green_asg_max}"
  min_size                  = "${var.green_asg_min}"
  desired_capacity          = "${var.green_asg_desired}"
  health_check_grace_period = 300
  health_check_type         = "ELB"
  launch_configuration      = "${module.launch_config.name}"
  target_group_arns         = ["${aws_alb_target_group.application_target_group.arn}"]
  vpc_zone_identifier       = ["${data.aws_subnet.application_subnet.*.id}"]
  enabled_metrics           = [
    "GroupMinSize", "GroupMaxSize", "GroupDesiredCapacity",
    "GroupInServiceInstances", "GroupPendingInstances", "GroupStandbyInstances",
    "GroupTerminatingInstances", "GroupTotalInstances"
  ]

  tags = [
    {
      key                 = "Name"
      value               = "${var.application_name}-${var.environment}"
      propagate_at_launch = true
    },
    {
      key                 = "app"
      value               = "${var.application_name}"
      propagate_at_launch = true
    },
    {
      key                 = "env"
      value               = "${var.environment}"
      propagate_at_launch = true
    },
    {
      key                 = "release"
      value               = "latest"
      propagate_at_launch = true
    }
  ]

  lifecycle {
    ignore_changes = ["launch_configuration"]
  }
}

# Autoscaling policies
resource "aws_autoscaling_policy" "highcpu" {
    name = "${var.application_name}-${var.environment}-high-cpu-scaleup"
    scaling_adjustment = 2
    adjustment_type = "ChangeInCapacity"
    cooldown = 300
    autoscaling_group_name = "${aws_autoscaling_group.asg_green.name}"
}

resource "aws_cloudwatch_metric_alarm" "highcpu" {
    alarm_name = "${var.application_name}-${var.environment}-high-cpu"
    comparison_operator = "GreaterThanOrEqualToThreshold"
    evaluation_periods = "2"
    metric_name = "CPUUtilization"
    namespace = "AWS/EC2"
    period = "120"
    statistic = "Average"
    threshold = "60"
    dimensions {
        AutoScalingGroupName = "${aws_autoscaling_group.asg_green.name}"
    }
    alarm_description = "Watch CPU usage for ${aws_autoscaling_group.asg_green.name} ASG"
    alarm_actions = ["${aws_autoscaling_policy.highcpu.arn}"]
}

resource "aws_autoscaling_policy" "lowcpu" {
    name = "${var.application_name}-${var.environment}-low-cpu-scaledown"
    scaling_adjustment = -1
    adjustment_type = "ChangeInCapacity"
    cooldown = 300
    autoscaling_group_name = "${aws_autoscaling_group.asg_green.name}"
}

resource "aws_cloudwatch_metric_alarm" "lowcpu" {
    alarm_name = "${var.application_name}-${var.environment}-low-cpu"
    comparison_operator = "LessThanOrEqualToThreshold"
    evaluation_periods = "2"
    metric_name = "CPUUtilization"
    namespace = "AWS/EC2"
    period = "120"
    statistic = "Average"
    threshold = "20"
    dimensions {
        AutoScalingGroupName = "${aws_autoscaling_group.asg_green.name}"
    }
    alarm_description = "Watch CPU usage for ${aws_autoscaling_group.asg_green.name} ASG"
    alarm_actions = ["${aws_autoscaling_policy.lowcpu.arn}"]
}
