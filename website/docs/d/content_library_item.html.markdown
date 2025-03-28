---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_content_library_item"
sidebar_current: "docs-vcfa-data-source-content-library-item"
description: |-
  Provides a data source to read a Content Library Item in VMware Cloud Foundation Automation. This can be used to obtain the details
  of Content Library Items, such as description, creation date, subscription details, etc.
---

# vcfa\_content\_library\_item

Provides a data source to read a Content Library Item in VMware Cloud Foundation Automation. This can be used to obtain the details
of Content Library Items, such as description, creation date, subscription details, etc.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_org" "system" {
  name = "System"
}

# It is a PROVIDER Content Library
data "vcfa_content_library" "cl" {
  org_id = data.vcfa_org.system.id
  name   = "My Library"
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
* `content_library_id` - (Required) ID of the [Content Library][vcfa_content_library-ds] that this item belongs to

## Attribute reference

All arguments and attributes defined in [`vcfa_content_library_item`][vcfa_content_library_item] resource are supported
as read-only (Computed) values.

[vcfa_content_library]: /providers/vmware/vcfa/latest/docs/data-sources/content_library
[vcfa_content_library-ds]: /providers/vmware/vcfa/latest/docs/data-sources/content_library
[vcfa_content_library_item]: /providers/vmware/vcfa/latest/docs/resources/content_library_item