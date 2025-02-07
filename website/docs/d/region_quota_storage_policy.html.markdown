---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_region_quota_storage_policy"
sidebar_current: "docs-vcfa-data-source-region-quota-storage-policy"
description: |-
  Provides a data source to read Storage Policies of Organization Region Quotas in VMware Cloud Foundation Automation.
---

# vcfa\_region\_quota\_storage\_policy

Provides a data source to read Storage Policies of Organization Region Quotas in VMware Cloud Foundation Automation.

## Example Usage

```hcl
```

## Argument Reference

The following arguments are supported:

- `org_region_quota_id` - (Required) Parent [Region Quota](/providers/vmware/vcfa/latest/docs/data-sources/org_region_quota) ID
- `name` - The name of the . It follows RFC 1123 Label Names to conform with Kubernetes standards


## Attribute Reference

All the arguments and attributes defined in
[`vcfa_region_quota_storage_policy`](/providers/vmware/vcfa/latest/docs/resources/region_quota_storage_policy) resource are available.