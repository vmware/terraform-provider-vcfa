---
page_title: "VMware Cloud Foundation Automation: vcfa_shared_subnet"
subcategory: ""
description: |-
  Provides a data source to read a Shared Subnet in VMware Cloud Foundation Automation.
---

# vcfa_shared_subnet

Provides a data source to read a Shared Subnet in VMware Cloud Foundation Automation.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "demo-region"
}

data "vcfa_shared_subnet" "demo" {
  name      = "demo-shared-subnet"
  region_id = data.vcfa_region.demo.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the Shared Subnet
- `region_id` - (Required) The ID of the Region that has this Shared Subnet. Can be looked up using
  [`vcfa_region`](/providers/vmware/vcfa/latest/docs/data-sources/region)

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_shared_subnet`](/providers/vmware/vcfa/latest/docs/resources/shared_subnet) resource are available.
