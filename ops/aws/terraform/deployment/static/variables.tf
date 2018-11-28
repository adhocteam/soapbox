variable "environment" {
  type = "string"
}

variable "application_name" {
  type = "string"
}

variable "region" {
  type    = "string"
  default = "us-east-1"
}

variable "not-found-response-path" {
  default = "/404.html"
}

variable "trusted_signers" {
  type = "list"
  default = []
}

variable "platform_domain" {
  type    = "string"
  default = "soapbox.hosting"
}

variable "tags" {
  type        = "map"
  description = "Optional Tags"
  default     = {}
}
