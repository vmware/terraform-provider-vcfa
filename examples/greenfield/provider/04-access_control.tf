# Create some Rights Bundles and Roles

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/rights_bundle
resource "vcfa_rights_bundle" "example" {
  name        = "tf-example-rights-bundle"
  description = "Created by Terraform VCFA Provider"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = false
  org_ids = [
    vcfa_org.example.id
  ]
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/role
resource "vcfa_role" "example-role" {
  org_id      = vcfa_org.example.id
  name        = "tf-example-role"
  description = "Created by Terraform VCFA Provider"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/global_role
resource "vcfa_global_role" "new-global-role" {
  name        = "tf-example-global-role"
  description = "Created by Terraform VCFA Provider"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = false
  org_ids = [
    vcfa_org.example.id
  ]
}