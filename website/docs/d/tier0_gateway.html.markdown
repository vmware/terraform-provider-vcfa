---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_tier0_gateway"
sidebar_current: "docs-vcfa-data-source-tier0-gateway"
description: |-
  Provides a VMware Cloud Foundation Automation Tier 0 Gateway data source.
---

# vcfa\_tier0\_gateway

Provides a VMware Cloud Foundation Automation Tier 0 Gateway data source.

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "region-one"
}

data "vcfa_tier0_gateway" "demo" {
  name      = "my-tier0-gateway"
  region_id = data.vcfa_region.demo.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of TM Tier 0 Gateway originating in NSX-T 
* `region_id` - (Required) An ID of Region. Can be looked up using
  [vcfa_region](/providers/vmware/vcfa/latest/docs/data-sources/region) data source

## Attribute Reference

* `description` - Description of the Tier 0 Gateway
* `parent_tier_0_id` - Parent Tier 0 Gateway ID if this is a Tier 0 VRF
* `already_imported` - Boolean flag if the Tier 0 Gateway is already consumed
