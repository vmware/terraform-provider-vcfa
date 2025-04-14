# Create Content Libraries

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/storage_class
data "vcfa_storage_class" "example" {
  region_id = data.vcfa_region.example.id
  name      = var.storage_class
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/content_library
resource "vcfa_content_library" "tenant-example" {
  org_id            = data.vcfa_org.example.id
  name              = "tf-tenant-example-content-library"
  description       = "Created by Terraform VCFA Provider"
  storage_class_ids = [data.vcfa_storage_class.example.id]
}