variable "region" {
  type    = "string"
  default = "us-east-1"
}

variable "application_state_bucket" {
  type    = "string"
  default = "soapbox-application-tf-state"
}

variable "application_state_table" {
  type    = "string"
  default = "soapbox-application-state-locking"
}

variable "application_releases_bucket" {
  type    = "string"
  default = "soapbox-application-releases"
}
