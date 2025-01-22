---
layout: "vcd"
page_title: "VMware Cloud Director: vcfa_org_vdc"
sidebar_current: "docs-vcd-datasource-tm-org-vdc"
description: |-
  Provides a data source to manage VMware Cloud Foundation Tenant Manager Organization VDC.
---

# vcd\_tm\_org\_vdc

Provides a data source to manage VMware Cloud Foundation Tenant Manager Organization VDC (Region Quota).

## Example Usage

```hcl
data "vcfa_org" "org" {
  name = "my-org"
}

data "vcfa_region" "region" {
  name = "region-one"
}

data "vcfa_org_vdc" "test" {
  org_id    = data.vcfa_org.org.id
  region_id = data.vcfa_region.region.id
}
```

## Argument Reference

The following arguments are supported:

* `region_id` - (Required)  An ID for the parent Region
* `org_id` - (Required) An ID for the parent Organization

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_vdc`](/providers/vmware/vcd/latest/docs/resources/tm_org_vdc) resource are available.
