---
page_title: "VMware Cloud Foundation Automation: vcfa_org_settings"
subcategory: ""
description: |-
  Provides a resource to manage the General Settings of an Organization in VMware Cloud Foundation Automation.
---

# vcfa_org_settings

Provides a resource to manage the General Settings of an [Organization][vcfa_org] in VMware Cloud Foundation Automation.

_Used by: **Provider**_

-> For Organization Networking settings, see [`vcfa_org_networking`](/providers/vmware/vcfa/latest/docs/resources/org_networking) resource.

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

- `org_id` - (Required) An [Organization](/providers/vmware/vcfa/latest/docs/data-sources/organization) ID for which the settings are to be changed
- `can_create_subscribed_libraries` - (Required) Whether the Organization can create [Content Libraries](/providers/vmware/vcfa/latest/docs/resources/content_library) that are subscribed to external sources
- `quarantine_content_library_items` - (Required) Whether to quarantine new [Content Library Items](/providers/vmware/vcfa/latest/docs/resources/content_library_item) for file inspection

~> Be careful as `quarantine_content_library_items=true` will make all the [`vcfa_content_library_item`](/providers/vmware/vcfa/latest/docs/resources/content_library_item) uploads for that
Organization to be blocked, waiting for manual upload approval

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Org configuration can be [imported][docs-import] into this resource via supplying path
for it. An example is below:

```
terraform import vcfa_org_settings.imported my-org-name
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-org-name` Organization settings.

After that, you can expand the configuration file and either update or delete the Organization settings as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Organization settings' stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org