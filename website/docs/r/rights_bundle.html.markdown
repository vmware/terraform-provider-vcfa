---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_rights_bundle"
sidebar_current: "docs-vcfa-resource-rights-bundle"
description: |-
 Provides a VMware Cloud Foundation Automation Rights Bundle resource. This can be used to create, modify, and delete Rights Bundles.
---

# vcfa\_rights\_bundle

Provides a VMware Cloud Foundation Automation Rights Bundle resource. This can be used to create, modify, and delete Rights Bundles.

## Example Usage

```hcl
resource "vcfa_rights_bundle" "new-rights-bundle" {
  name        = "new-rights-bundle"
  description = "new rights bundle from CLI"
  rights = [
    "Catalog: Add vApp from My Cloud",
    "Catalog: Edit Properties",
    "Catalog: View Private and Shared Catalogs",
    "Organization vDC Compute Policy: View",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
  ]
  publish_to_all_tenants = false
  tenants = [
    "org1",
    "org2",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the rights bundle.
* `description` - (Required) A description of the rights bundle
* `rights` - (Optional) Set of rights assigned to this role
* `publish_to_all_tenants` - (Required) When true, publishes the rights bundle to all tenants
* `tenants` - (Optional) Set of tenants to which this rights bundle gets published. Ignored if `publish_to_all_tenants` is true.

## Attribute Reference

* `read_only` - Whether this rights bundle is read-only
* `bundle_key` - Key used for internationalization

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing rights bundle can be [imported][docs-import] into this resource via supplying the rights bundle name (the rights
bundle is at the top of the entity hierarchy).
For example, using this structure, representing an existing rights bundle that was **not** created using Terraform:

```hcl
resource "vcfa_rights_bundle" "default-set" {
  name = "Default Rights Bundle"
}
```

You can import such rights bundle into terraform state using this command

```
terraform import vcfa_rights_bundle.default-set "Default Rights Bundle"
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCFA_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the rights bundle as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the rights bundle's stored properties.
