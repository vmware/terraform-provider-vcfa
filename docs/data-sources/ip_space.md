---
page_title: "VMware Cloud Foundation Automation: vcfa_ip_space"
subcategory: ""
description: |-
  Provides a data source to read an IP Space in VMware Cloud Foundation Automation.
---

# vcfa_ip_space

Provides a data source to read an IP Space in VMware Cloud Foundation Automation.

_Used by: **Provider**_

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

- `name` - (Required) The name of the IP Space
- `region_id` - (Required) The ID of the Region that has this IP Space definition. Can be looked up using
  [`vcfa_region`](/providers/vmware/vcfa/latest/docs/data-sources/region)

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_ip_space`](/providers/vmware/vcfa/latest/docs/resources/ip_space) resource are available.
