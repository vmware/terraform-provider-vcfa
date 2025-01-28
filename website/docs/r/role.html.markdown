---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_role"
sidebar_current: "docs-vcfa-resource-role"
description: |-
 Provides a VMware Cloud Foundation Automation Role. This can be used to create, modify, and delete Roles.
---

# vcfa\_role

Provides a VMware Cloud Foundation Automation Role. This can be used to create, modify, and delete Roles.

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
    "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
    "Catalog: View Private and Shared Catalogs",
    "Organization vDC Compute Policy: View",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) The ID of organization to use
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
resource "vcfa_role" "catalog-author" {
  org_id = "urn:vcloud:org:a93c9db9-7471-3192-8d09-a8f7eeda85f9"
  name   = "Catalog Author"
}
```

You can import such Role into terraform state using this command

```
terraform import vcfa_role.catalog-author "my-org.Catalog Author"
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCFA_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the Role as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Role's stored properties.
