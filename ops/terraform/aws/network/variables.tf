variable "environment" {
  type = "string"
}

variable "application_name" {
  type = "string"
}

variable "application_domain" {
  type = "string"
}

variable "instance_tenancy" {
  type = "string"
  default = "default"
}

variable "region" {
  type    = "string"
  default = "us-east-1"
}

variable "availability_zones" {
  type = "list"
  default = ["a", "b"]
}

variable "public_ingress_cidrs" {
  type    = "list"
  default = ["0.0.0.0/0"]
}

variable "vpc_cidr_block" {
  type    = "string"
  default = "10.0.0.0/16"
}

variable "az_cidr_blocks" {
  type = "map"
  default = {
    "app"  = ["10.0.0.0/20", "10.0.16.0/20"]
    "dmz"  = ["10.0.32.0/20", "10.0.48.0/20"]
  }
}
