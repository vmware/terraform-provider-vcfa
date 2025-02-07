---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_region_quota"
sidebar_current: "docs-vcfa-data-source-org-region-quota"
description: |-
  Provides a data source to read VMware Cloud Foundation Automation Organization Region Quotas.
---

# vcfa\_org\_region\_quota

Provides a data source to read VMware Cloud Foundation Automation Organization Region Quotas.

## Example Usage

```hcl
data "vcfa_org" "org" {
  name = "my-org"
}

data "vcfa_region" "region" {
  name = "region-one"
}

data "vcfa_org_region_quota" "test" {
  org_id    = data.vcfa_org.org.id
  region_id = data.vcfa_region.region.id
}
```

## Argument Reference

The following arguments are supported:

- `region_id` - (Required)  An ID for the parent Region
- `org_id` - (Required) An ID for the parent Organization

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_region_quota`](/providers/vmware/vcfa/latest/docs/resources/org_region_quota) resource are available.
