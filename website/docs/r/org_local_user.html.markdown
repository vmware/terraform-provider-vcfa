---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_local_user"
sidebar_current: "docs-vcfa-resource-org-local-user"
description: |-
  Provides a resource to manage local users in VMware Cloud Foundation Automation Organizations.
---

# vcfa\_org\_local\_user

Provides a resource to manage local users in VMware Cloud Foundation Automation Organizations.

## Example Usage

```hcl
resource "vcfa_org" "demo" {
  name         = "terraform-org"
  display_name = "terraform-org"
  description  = "Terraform demo"
  is_enabled   = true
}

data "vcfa_role" "org-admin" {
  org_id = vcfa_org.demo.id
  name   = "Organization Administrator"
}

resource "vcfa_org_local_user" "demo" {
  org_id   = vcfa_org.demo.id
  role_id  = data.vcfa_role.org-admin.id
  username = "demo-local-user"
  password = "CHANGE-ME"
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) An Org ID for this Local User to be created in 
- `role_id` - (Required) A role ID to assign to this user
- `username` - (Required) User name for this local user
- `password` - (Required) A password for the user

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Org configuration can be [imported][docs-import] into this resource via supplying path
for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcfa_org_local_user.imported my-org-name.my-user-name
```

The above would import the `my-user-name` local user from  `my-org-name` Organization.
