# Create Content Libraries

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/storage_class
data "vcfa_storage_class" "example" {
  for_each  = var.vcenter_storage_policy_names
  region_id = vcfa_region.example.id
  name      = each.key
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/content_library
resource "vcfa_content_library" "example" {
  org_id      = data.vcfa_org.system.id
  name        = "tf-example-content-library"
  description = "Created by Terraform VCFA Provider"
  storage_class_ids = [
    values(data.vcfa_storage_class.example)[0].id
  ]
}
