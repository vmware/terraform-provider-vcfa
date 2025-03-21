---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_content_library"
sidebar_current: "docs-vcfa-data-source-content-library"
description: |-
  Provides a data source to read a Content Library in VMware Cloud Foundation Automation. This can be used to obtain the details
  of Content Libraries, such as description, creation date, etc.
---

# vcfa\_content\_library

Provides a data source to read a Content Library in VMware Cloud Foundation Automation. This can be used to obtain the details
of Content Libraries, such as description, creation date, etc.

-> This data source can be used by both **System Administrators** and **Tenant users**

## Example Usage for Provider libraries

```hcl
data "vcfa_org" "system" {
  name = "System"
}

data "vcfa_content_library" "cl" {
  org_id = data.vcfa_org.system.id
  name   = "My Library"
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
* `org_id` - (Required) The reference to the Organization that the Content Library belongs to. For Content Libraries of type `PROVIDER`,
  a reference to the `System` org with [`vcfa_org` data source](/providers/vmware/vcfa/latest/docs/data-sources/org) must be provided

## Attribute reference

All arguments and attributes defined in [`vcfa_content_library` resource](/providers/vmware/vcfa/latest/docs/resources/content_library) are supported
as read-only (Computed) values.
