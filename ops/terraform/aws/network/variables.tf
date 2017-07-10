variable "default_region" {
  type    = "string"
  default = "us-east-1"
}

variable "environment" {
  type    = "string"
  default = "production"
}

variable "application_name" {
  type    = "string"
  default = "A Soapbox-managed Application"
}

variable "application_domain" { type = "string" }

variable "application_port" {
  type    = "string"
  default = "80"
}

variable "pubic_ingress_cidrs" {
  type    = "list"
  default = ["0.0.0.0/0"]
}

variable "vpc_cidr_block" {
  type    = "string"
  default = "10.0.0.0/16"
}
