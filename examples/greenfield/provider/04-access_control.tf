# Create some Rights Bundles and Roles

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/rights_bundle
resource "vcfa_rights_bundle" "demo" {
  name        = "tf-demo-rights-bundle"
  description = "Created by Terraform VCFA Provider"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = false
  org_ids = [
    vcfa_org.demo.id
  ]
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/role
resource "vcfa_role" "demo-role" {
  org_id      = vcfa_org.demo.id
  name        = "tf-demo-role"
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
  name        = "tf-demo-global-role"
  description = "Created by Terraform VCFA Provider"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = false
  org_ids = [
    vcfa_org.demo.id
  ]
}