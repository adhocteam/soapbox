variable "environment" {
  type = "string"
}

variable "application_name" {
  type = "string"
}

variable "release" {
  type    = "string"
  default = "latest"
}

variable "region" {
  type    = "string"
  default = "us-east-1"
}

variable "availability_zones" {
  type = "list"
  default = ["a", "b"]
}

variable "green_asg_max" {
  type = "string"
  default = 6
}

variable "green_asg_min" {
  type = "string"
  default = 1
}

variable "green_asg_desired" {
  type = "string"
  default = 1
}

variable "application_acm_cert_arn" {
  type = "string"
  default = ""
}

variable "application_port" {
  type    = "string"
  default = "80"
}

variable "health_check_path" {
  type = "string"
  default = "/_health"
}
