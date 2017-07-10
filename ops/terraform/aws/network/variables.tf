variable "environment" {
  type = "string"
}

variable "application_name" {
  type = "string"
}

variable "application_domain" {
  type = "string"
}

variable "region" {
  type    = "string"
  default = "us-east-1"
}

variable "availability_zones" {
  type = "list"
  default = ["a", "b"]
}

variable "application_port" {
  type    = "string"
  default = "80"
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
    "app"  = ["10.0.1.0/24", "10.0.2.0/24"]
    "dmz"  = ["10.0.3.0/24", "10.0.4.0/24"]
  }
}
