---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_role"
sidebar_current: "docs-vcfa-resource-role"
description: |-
  Provides a resource to manage Roles in VMware Cloud Foundation Automation.
---

# vcfa\_role

Provides a resource to manage Roles in VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_org" "org1" {
  name = "org1"
}

resource "vcfa_role" "new-role" {
  org_id      = data.vcfa_org.org1.id
  name        = "new-role"
  description = "New role from Terraform"
  rights = [
    "Content Library Item: Manage",
    "Content Library Item: View",
    "Content Library: Manage",
    "Content Library: View",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) The ID of Organization that will have the Role. Can be fetched with [`vcfa_org` data source](/providers/vmware/vcfa/latest/docs/data-sources/org)
* `name` - (Required) The name of the Role
* `description` - (Required) A description of the Role
* `rights` - (Optional) Set of rights assigned to this Role

## Attribute Reference

* `read_only` - Whether this Role is read-only
* `bundle_key` - Key used for internationalization

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Role can be [imported][docs-import] into this resource via supplying the full dot separated path for a Role.
For example, using this structure, representing an existing Role that was **not** created using Terraform:

```hcl
data "vcfa_org" "org1" {
  name = "my-org"
}

resource "vcfa_role" "my-existing-role" {
  org_id = data.vcfa_org.org1.id
  name   = "Blueprint Publisher"
}
```

You can import such Role into terraform state using this command

```
terraform import vcfa_role.my-existing-role "my-org.Blueprint Publisher"
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

After that, you can expand the configuration file and either update or delete the Role as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Role's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources