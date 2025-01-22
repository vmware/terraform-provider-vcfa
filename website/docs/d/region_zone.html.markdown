---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_region_zone"
sidebar_current: "docs-vcfa-data-source-region-zone"
description: |-
  Provides a data source to read VMware Cloud Foundation Automation Region Zone that can be used when creating a VDC (Region Quota).
---

# vcfa\_region\_zone

Provides a data source to read VMware Cloud Foundation Automation Region Zone that can be used when creating a VDC (Region Quota).

## Example Usage

```hcl
data "vcfa_region" "region" {
  name = "region-one"
}

data "vcfa_region_zone" "my" {
  region_id = vcfa_region.region.id
  name      = "my-zone"
}
```

## Argument Reference

The following arguments are supported:

- `region_id` - (Required) Parent Region ID
- `name` - (Required) Name of Region Zone

## Attribute Reference

- `cpu_limit_mhz` - Total amount of reserved and unreserved CPU resources allocated in MHz
- `cpu_reservation_mhz` - Total amount of CPU resources reserved in MHz
- `cpu_reservation_used_mhz` - The amount of CPU resources used in MHz. For Tenants, this value
  represents the total given to all of a Tenant's Namespaces. For Providers, this value represents
  the total given to all Tenants
- `memory_limit_mib` - Total amount of reserved and unreserved memory resources allocated in MiB
- `memory_reservation_used_mib` - Total amount of reserved memory resources used in MiB. For
  Tenants, this value represents the total given to all of a Tenant's Namespaces. For Providers,
  this value represents the total given to all Tenants
- `memory_reservation_mib` - Total amount of reserved and unreserved memory resources used in MiB.
  For Tenants, this value represents the total given to all of a Tenant's Namespaces. For Providers,
  this value represents the total given to all Tenants
