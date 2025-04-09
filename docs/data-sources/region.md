---
page_title: "VMware Cloud Foundation Automation: vcfa_region"
subcategory: ""
description: |-
  Provides a data source to read a Region in VMware Cloud Foundation Automation.
---

# vcfa_region

Provides a data source to read a Region in VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_region" "one" {
  name = "region-one"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) A name of existing Region

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_region`](/providers/vmware/vcfa/latest/docs/resources/region) resource are available.
