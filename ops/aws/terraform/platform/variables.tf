variable "region" {
  type    = "string"
  default = "us-east-1"
}

variable "application_releases_bucket" {
  type    = "string"
  default = "soapbox-app-images"
}

variable "application_state_bucket" {
  type    = "string"
  default = "soapbox-app-tf-state"
}

variable "application_state_table" {
  type    = "string"
  default = "soapbox-app-state-locking"
}

variable "platform_domain" {
  type    = "string"
  default = "soapbox.hosting"
}
