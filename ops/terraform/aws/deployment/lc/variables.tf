variable "environment" {
  type = "string"
}

variable "application_name" {
  type = "string"
}

variable "release" {
  type = "string"
}

variable "region" {
  type    = "string"
  default = "us-east-1"
}

variable "instance_type" {
  type    = "string"
  default = "t2.micro"
}
