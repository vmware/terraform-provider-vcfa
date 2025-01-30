---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_global_role"
sidebar_current: "docs-vcfa-resource-global-role"
description: |-
 Provides a VMware Cloud Foundation Automation Global Role. This can be used to create, modify, and delete Global Roles.
---

# vcfa\_global\_role

Provides a VMware Cloud Foundation Automation Global Role. This can be used to create, modify, and delete Global Roles.

## Example Usage

```hcl
data "vcfa_org" "org1" {
  name = "org1"
}

data "vcfa_org" "org2" {
  name = "org2"
}

resource "vcfa_global_role" "new-global-role" {
  name        = "new-global-role"
  description = "New Global Role from Terraform"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = false
  org_ids = [
    data.vcfa_org.org1.id,
    data.vcfa_org.org2.id,
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Global Role.
* `description` - (Required) A description of the Global Role
* `rights` - (Optional) List of rights assigned to this Global Role
* `publish_to_all_orgs` - (Required) When true, publishes the Global Role to all Organizations
* `org_ids` - (Optional) List of IDs of the Organizations to which this Global Role gets published. Ignored if `publish_to_all_orgs` is `true`

## Attribute Reference

* `read_only` - Whether this Global Role is read-only
* `bundle_key` - Key used for internationalization

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Global Role can be [imported][docs-import] into this resource via supplying the Global Role name (the global
role is at the top of the entity hierarchy).
For example, using this structure, representing an existing Global Role that was **not** created using Terraform:

```hcl
resource "vcfa_global_role" "my-global-role" {
  name = "My Existing Role"
}
```

You can import such Global Role into terraform state using this command

```
terraform import vcfa_global_role.my-global-role "My Existing Role"
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCFA_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the Global Role as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Global Role's stored properties.
