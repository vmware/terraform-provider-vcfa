---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_storage_class"
sidebar_current: "docs-vcfa-data-source-storage-class"
description: |-
  Provides a VMware Cloud Foundation Automation data source to read Storage Classes.
---

# vcfa\_storage\_class

Provides a VMware Cloud Foundation Automation data source to read Storage Classes.

## Example Usage

```hcl
data "vcfa_region" "region" {
  name = "my-region"
}

data "vcfa_region_storage_class" "sc" {
  region_id = data.vcfa_region.region.id
  name      = "vSAN Default Storage Class"
}

resource "vcfa_content_library" "cl" {
  name        = "My Library"
  description = "A simple library"
  storage_class_ids = [
    data.vcfa_storage_class.sc.id
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Storage Class to read
* `region_id` - (Required) The ID of the Region where the Storage Class belongs

## Attribute reference

* `storage_capacity_mib` - The total storage capacity of the Storage Class in mebibytes
* `storage_consumed_mib` - For tenants, this represents the total storage given to all namespaces consuming from this
  Storage Class in mebibytes. For providers, this represents the total storage given to tenants from this Storage Class
  in mebibytes
* `zone_ids` - A set with all the IDs of the zones available to the Storage Class