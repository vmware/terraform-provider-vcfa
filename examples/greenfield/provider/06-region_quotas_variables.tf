variable "region_vm_class_names" {
  type        = set(string)
  description = "Names of VM Classes"
}

variable "region_storage_policy_names" {
  type        = set(string)
  description = "Region storage policies"
}
