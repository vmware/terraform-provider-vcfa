---
page_title: "VMware Cloud Foundation Automation: vcfa_provider_gateway"
description: |-
  Provides a data source to read a Provider Gateway in VMware Cloud Foundation Automation.
---

# vcfa_provider_gateway

Provides a data source to read a Provider Gateway in VMware Cloud Foundation Automation.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "region-one"
}

data "vcfa_provider_gateway" "demo" {
  name      = "Demo Provider Gateway"
  region_id = data.vcfa_region.demo.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of Provider Gateway
- `region_id` - (Required) An ID of Region. Can be looked up using
  [vcfa_region](/providers/vmware/vcfa/latest/docs/data-sources/region) data source


## Attribute Reference

All the arguments and attributes defined in
[`vcfa_provider_gateway`](/providers/vmware/vcfa/latest/docs/resources/provider_gateway)
resource are available.