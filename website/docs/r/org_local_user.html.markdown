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

data "vcfa_role" "org-user" {
  org_id = vcfa_org.demo.id
  name   = "Organization User"
}

resource "vcfa_org_local_user" "demo" {
  org_id   = vcfa_org.demo.id
  role_ids = [data.vcfa_role.org-admin.id, data.vcfa_role.org-user.id]
  username = "demo-local-user"
  password = "CHANGE-ME"
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) An Org ID for this Local User to be created in 
- `role_ids` - (Required) A set of role IDs to assign to this user
- `username` - (Required) User name for this local user
- `password` - (Required) A password for the user

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Org Local User configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

```
terraform import vcfa_org_local_user.imported my-org-name.my-user-name
```

NOTE: the default separator (.) can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-user-name` local user from  `my-org-name` Organization.

After that, you can expand the configuration file and either update or delete the Organization Local User as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Organization Local User's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources