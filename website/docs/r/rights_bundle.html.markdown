---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_rights_bundle"
sidebar_current: "docs-vcfa-resource-rights-bundle"
description: |-
 Provides a VMware Cloud Foundation Automation Rights Bundle resource. This can be used to create, modify, and delete Rights Bundles.
---

# vcfa\_rights\_bundle

Provides a VMware Cloud Foundation Automation Rights Bundle resource. This can be used to create, modify, and delete Rights Bundles.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_org" "org1" {
  name = "org1"
}

data "vcfa_org" "org2" {
  name = "org2"
}

resource "vcfa_rights_bundle" "new-rights-bundle" {
  name        = "new-rights-bundle"
  description = "new rights bundle from Terraform"
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

* `name` - (Required) The name of the Rights Bundle
* `description` - (Required) A description of the Rights Bundle
* `rights` - (Optional) Set of Rights assigned to this Rights Bundle
* `publish_to_all_orgs` - (Required) When `true`, publishes the Rights Bundle to all Organizations
* `org_ids` - (Optional) Set of IDs of the Organizations to which this Rights Bundle gets published. Ignored if `publish_to_all_orgs` is `true`

## Attribute Reference

* `read_only` - Whether this Rights Bundle is read-only
* `bundle_key` - Key used for internationalization

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Rights Bundle can be [imported][docs-import] into this resource via supplying the Rights Bundle name (the rights
bundle is at the top of the entity hierarchy).
For example, using this structure, representing an existing Rights Bundle that was **not** created using Terraform:

```hcl
resource "vcfa_rights_bundle" "default-set" {
  name = "Default Rights Bundle"
}
```

You can import such Rights Bundle into terraform state using this command

```
terraform import vcfa_rights_bundle.default-set "Default Rights Bundle"
```

NOTE: the default separator (.) can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

After that, you can expand the configuration file and either update or delete the Rights Bundle as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Rights Bundle's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources