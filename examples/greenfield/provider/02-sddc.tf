# Configure underlying vCenter and NSX

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/nsx_manager
resource "vcfa_nsx_manager" "demo" {
  name                   = "tf-demo-nsx-manager"
  description            = "Created by Terraform VCFA Provider"
  username               = var.nsx_manager_username
  password               = var.nsx_manager_password
  url                    = var.nsx_manager_url
  auto_trust_certificate = true
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/vcenter
resource "vcfa_vcenter" "demo" {
  nsx_manager_id             = vcfa_nsx_manager.demo.id
  name                       = "tf-demo-vcenter"
  description                = "Created by Terraform VCFA Provider"
  url                        = var.vcenter_url
  auto_trust_certificate     = true
  refresh_vcenter_on_create  = true
  refresh_policies_on_create = true
  username                   = var.vcenter_username
  password                   = var.vcenter_password
  is_enabled                 = true
}
