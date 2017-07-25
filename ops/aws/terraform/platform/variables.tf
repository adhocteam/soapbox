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
  default = "soapbox-app-state"
}

variable "application_state_locking_table" {
  type    = "string"
  default = "soapbox-app-state-locking"
}
