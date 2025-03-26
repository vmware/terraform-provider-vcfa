---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_ip_space"
sidebar_current: "docs-vcfa-data-source-ip-space"
description: |-
  Provides a VMware Cloud Foundation Automation IP Space data source.
---

# vcfa\_ip\_space

Provides a VMware Cloud Foundation Automation IP Space data source.

-> This data source can be used by both **System Administrators** and **Tenant users**

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "demo-region"
}

data "vcfa_ip_space" "demo" {
  name      = "demo-ip-space"
  region_id = data.vcfa_region.region.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of IP Space
* `region_id` - (Required) The Region ID that has this IP Space definition. Can be looked up using
  [`vcfa_region`](/providers/vmware/vcfa/latest/docs/data-sources/region)

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_ip_space`](/providers/vmware/vcfa/latest/docs/resources/ip_space) resource are available.
