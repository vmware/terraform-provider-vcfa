---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_content_library"
sidebar_current: "docs-vcfa-data-source-content-library"
description: |-
  Provides a VMware Cloud Foundation Automation Content Library data source. This can be used to read Content Libraries.
---

# vcfa\_content\_library

Provides a VMware Cloud Foundation Automation Content Library data source. This can be used to read Content Libraries.

## Example Usage for Provider libraries

```hcl
data "vcfa_content_library" "cl" {
  name = "My Library"
}

output "is_shared" {
  value = data.vcfa_content_library.cl.is_shared
}
output "owner_org" {
  value = data.vcfa_content_library.cl.org_id
}
```

## Example Usage for Tenant libraries

```hcl
data "vcfa_org" "my-org" {
  name = "my-org"
}

data "vcfa_content_library" "cl" {
  org_id = data.vcfa_org.my-org.id
  name   = "My Library"
}

output "is_shared" {
  value = data.vcfa_content_library.cl.is_shared
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Content Library to read
* `org_id` - (Optional) The reference to the Organization that the Content Library belongs to. If it is not set, assumes the
  Content Library is of type `PROVIDER`

## Attribute reference

All arguments and attributes defined in [the resource](/providers/vmware/vcfa/latest/docs/resources/content_library) are supported
as read-only (Computed) values.
