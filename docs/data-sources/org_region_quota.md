---
page_title: "VMware Cloud Foundation Automation: vcfa_org_region_quota"
subcategory: ""
description: |-
  Provides a data source to read a Region Quota from an Organization in VMware Cloud Foundation Automation.
---

# Data Source: vcfa_org_region_quota

Provides a data source to read a Region Quota from an [Organization][vcfa_org-ds] in VMware Cloud Foundation Automation.

_Used by: **Provider**_

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

- `region_id` - (Required)  An ID for the parent [Region][vcfa_region-ds]
- `org_id` - (Required) An ID for the parent [Organization][vcfa_org-ds]

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_region_quota`](/providers/vmware/vcfa/latest/docs/resources/org_region_quota) resource are available.

[vcfa_region-ds]: /providers/vmware/vcfa/latest/docs/data-sources/region
[vcfa_org-ds]: /providers/vmware/vcfa/latest/docs/data-sources/org
