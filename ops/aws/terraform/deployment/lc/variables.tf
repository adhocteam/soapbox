variable "environment" {
  type = "string"
}

variable "application_name" {
  type = "string"
}

variable "release" {
  type = "string"
  default = "latest"
}

variable "application_port" {
  type    = "string"
  default = "8080"
}

variable "region" {
  type    = "string"
  default = "us-east-1"
}

variable "instance_type" {
  type    = "string"
  default = "t2.micro"
}

variable "release_bucket" {
  type    = "string"
  default = "soapbox-app-images"
}

variable "key_name" {
  type    = "string"
  default = "soapbox-app"
}

variable "instance_profile" {
  type    = "string"
  default = "soapbox-app"
}
