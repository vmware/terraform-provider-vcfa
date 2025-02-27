---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_content_library_item"
sidebar_current: "docs-vcfa-data-source-content-library-item"
description: |-
  Provides a VMware Cloud Foundation Automation Content Library Item data source. This can be used to read Content Library Items.
---

# vcfa\_content\_library\_item

Provides a VMware Cloud Foundation Automation Content Library Item data source. This can be used to read Content Library Items.

## Example Usage

```hcl
data "vcfa_content_library" "cl" {
  name = "My Library"
}

data "vcfa_content_library_item" "cli" {
  name               = "My Library Item"
  content_library_id = data.vcfa_content_library.cl.id
}

output "is_published" {
  value = data.vcfa_content_library_item.cli.is_published
}
output "image_identifier" {
  value = data.vcfa_content_library_item.cli.image_identifier
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Content Library Item to read
* `content_library_id` - (Required) ID of the Content Library that this item belongs to. Can be obtained with [a data source](/providers/vmware/vcfa/latest/docs/data-sources/content_library)

## Attribute reference

All arguments and attributes defined in [the resource](/providers/vmware/vcfa/latest/docs/resources/content_library_item) are supported
as read-only (Computed) values.
