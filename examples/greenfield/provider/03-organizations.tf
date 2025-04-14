# Create some Organizations (tenants)

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org
resource "vcfa_org" "demo" {
  name         = "tf-demo-org"
  display_name = "tf-demo-org"
  description  = "Created by Terraform VCFA Provider"
  is_enabled   = true
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org_settings
resource "vcfa_org_settings" "demo" {
  org_id                           = vcfa_org.demo.id
  can_create_subscribed_libraries  = true
  quarantine_content_library_items = false
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/role
data "vcfa_role" "org-admin" {
  org_id = vcfa_org.demo.id
  name   = "Organization Administrator"
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org_local_user
resource "vcfa_org_local_user" "demo" {
  org_id   = vcfa_org.demo.id
  role_ids = [data.vcfa_role.org-admin.id]
  username = "tf-demo-local-user"
  password = "long-change-ME1"
}

# A classic VRA-style organization. See "is_classic_tenant" argument at:
# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org
resource "vcfa_org" "demo-classic" {
  name              = "tf-demo-org-classic"
  display_name      = "tf-demo-org-classic"
  description       = "Created by Terraform VCFA Provider"
  is_classic_tenant = true
  is_enabled        = true
}
