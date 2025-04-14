variable "url" {
  type        = string
  description = "VMware Cloud Foundation Automation URL, e.g. https://HOST"
}

variable "username" {
  type        = string
  description = "Username for authenticating"
}

variable "password" {
  type        = string
  sensitive   = true
  description = "Password for a given 'username'"
}

variable "org" {
  type        = string
  description = "Tenant name"
}

variable "region_name" {
  type        = string
  description = "Region name to use"
}
