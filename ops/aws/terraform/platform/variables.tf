variable "region" {
  type    = "string"
  default = "us-east-1"
}

variable "application_releases_bucket" {
  type    = "string"
  default = "soapbox-app-images"
}

variable "platform_domain" {
  type    = "string"
  default = "soapbox.hosting"
}
