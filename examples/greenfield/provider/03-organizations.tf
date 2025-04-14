# Create some Organizations (tenants)

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org
resource "vcfa_org" "example" {
  name         = "tf-example-org"
  display_name = "tf-example-org"
  description  = "Created by Terraform VCFA Provider"
  is_enabled   = true
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org_settings
resource "vcfa_org_settings" "example" {
  org_id                           = vcfa_org.example.id
  can_create_subscribed_libraries  = true
  quarantine_content_library_items = false
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/role
data "vcfa_role" "org-admin" {
  org_id = vcfa_org.example.id
  name   = "Organization Administrator"
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org_local_user
resource "vcfa_org_local_user" "example" {
  org_id   = vcfa_org.example.id
  role_ids = [data.vcfa_role.org-admin.id]
  username = "tf-example-local-user"
  password = "long-change-ME1"
}

# A classic VRA-style organization. See "is_classic_tenant" argument at:
# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org
resource "vcfa_org" "example-classic" {
  name              = "tf-example-org-classic"
  display_name      = "tf-example-org-classic"
  description       = "Created by Terraform VCFA Provider"
  is_classic_tenant = true
  is_enabled        = true
}
