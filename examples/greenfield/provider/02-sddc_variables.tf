variable "vcenter_url" {
  type        = string
  description = "URL of vCenter, e.g. https://HOST"
}

variable "vcenter_username" {
  type        = string
  description = "Username for authenticating to vCenter"
}

variable "vcenter_password" {
  type        = string
  sensitive   = true
  description = "Password for a given 'vcenter_username'"
}

variable "nsx_manager_url" {
  type        = string
  description = "URL of NSX manager, e.g. https://HOST"
}

variable "nsx_manager_username" {
  type        = string
  description = "Username for authenticating to NSX Manager"
}

variable "nsx_manager_password" {
  type        = string
  sensitive   = true
  description = "Password for a given 'nsx_manager_username'"
}
