variable "supervisor_name" {
  type        = string
  description = "Name of Supervisor in vCenter"
}

variable "supervisor_zone_name" {
  type        = string
  description = "Name of Supervisor Zone in vCenter"
}

variable "vcenter_storage_policy_names" {
  type        = set(string)
  description = "vCenter storage profiles"
}