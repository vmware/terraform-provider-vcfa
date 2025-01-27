---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_networking"
sidebar_current: "docs-vcfa-resource-org-networking"
description: |-
  Provides a resource to manage VMware Cloud Foundation Automation Organization Networking Settings.
---

# vcfa\_org\_networking

Provides a resource to manage VMware Cloud Foundation Automation Organization Networking Settings.

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "my-org-name"
}

resource "vcfa_org_networking" "demo" {
  org_id   = data.vcfa_org.demo.id
  log_name = "org1"
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) An Organization ID for which the networking settings are to be changed
- `log_name` - (Required) A globally unique identifier for this Organization in the logs of the
  backing network provider. Must be 1-8 chars length.


## Attribute Reference

The following attributes are exported on this resource:

- `networking_tenancy_enabled` - Whether this Organization has tenancy for the network domain in the
  backing network provider

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Org configuration can be [imported][docs-import] into this resource via supplying path
for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcfa_org_networking.imported my-org-name
```

The above would import the `my-org-name` Organization Networking settings.