---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_settings"
sidebar_current: "docs-vcfa-resource-org-settings"
description: |-
  Provides a resource to manage VMware Cloud Foundation Automation Organization general Settings.
---

# vcfa\_org\_settings

Provides a resource to manage VMware Cloud Foundation Automation Organization general Settings.

-> For Organization Networking settings, see [`vcfa_org_networking`](/providers/vmware/vcfa/latest/docs/resources/org_networking) resource

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "my-org-name"
}

resource "vcfa_org_settings" "demo" {
  org_id                           = data.vcfa_org.demo.id
  can_create_subscribed_libraries  = true
  quarantine_content_library_items = true
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) An Organization ID for which the settings are to be changed
- `can_create_subscribed_libraries` - (Required) Whether the Organization can create content libraries that are subscribed to external sources
- `quarantine_content_library_items` - (Required) Whether to quarantine new Content Library Items for file inspection


## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Org configuration can be [imported][docs-import] into this resource via supplying path
for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcfa_org_settings.imported my-org-name
```

The above would import the `my-org-name` Organization settings.