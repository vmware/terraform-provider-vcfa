---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_content_library"
sidebar_current: "docs-vcfa-data-source-tm-content-library"
description: |-
  Provides a VMware Cloud Foundation Automation Content Library data source. This can be used to read Content Libraries.
---

# vcfa\_content\_library

Provides a VMware Cloud Foundation Automation Content Library data source. This can be used to read Content Libraries.

This data source is exclusive to **VMware Cloud Foundation Automation**. Supported in provider *v4.0+*

## Example Usage

```hcl
data "vcfa_content_library" "cl" {
  name = "My Library"
}

output "is_shared" {
  value = data.vcfa_content_library.cl.is_shared
}
output "owner_org" {
  value = data.vcfa_content_library.cl.owner_org_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Content Library to read

## Attribute reference

All arguments and attributes defined in [the resource](/providers/vmware/vcfa/latest/docs/resources/content_library) are supported
as read-only (Computed) values.
