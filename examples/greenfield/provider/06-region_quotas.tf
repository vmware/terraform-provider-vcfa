# Create a Region Quota to be able to start deploying workloads and creating Content Libraries

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/region_zone
data "vcfa_region_zone" "example" {
  region_id = vcfa_region.example.id
  name      = var.supervisor_zone_name
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/region_vm_class
data "vcfa_region_vm_class" "example" {
  for_each  = var.region_vm_class_names
  region_id = vcfa_region.example.id
  name      = each.key
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/region_storage_policy
data "vcfa_region_storage_policy" "example" {
  for_each  = var.region_storage_policy_names
  name      = each.key
  region_id = vcfa_region.example.id
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org_region_quota
resource "vcfa_org_region_quota" "example" {
  org_id         = vcfa_org.example.id
  region_id      = vcfa_region.example.id
  supervisor_ids = [data.vcfa_supervisor.example.id]
  zone_resource_allocations {
    region_zone_id         = data.vcfa_region_zone.example.id
    cpu_limit_mhz          = 2000
    cpu_reservation_mhz    = 100
    memory_limit_mib       = 1024
    memory_reservation_mib = 512
  }
  region_vm_class_ids = values(data.vcfa_region_vm_class.example)[*].id
  dynamic "region_storage_policy" {
    for_each = values(data.vcfa_region_storage_policy.example)
    content {
      region_storage_policy_id = region_storage_policy.value["id"]
      storage_limit_mib        = 1024
    }
  }
}