---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_vdc"
sidebar_current: "docs-vcfa-data-source-org-vdc"
description: |-
  Provides a data source to manage VMware Cloud Foundation Automation Organization VDC.
---

# vcfa\_org\_vdc

Provides a data source to manage VMware Cloud Foundation Automation Organization VDC (Region Quota).

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
[`vcfa_org_vdc`](/providers/vmware/vcfa/latest/docs/resources/org_vdc) resource are available.
