---
page_title: "VMware Cloud Foundation Automation: vcfa_region_storage_policy"
subcategory: ""
description: |-
  Provides a data source to read a Region Storage Policy in VMware Cloud Foundation Automation. These are useful
  when configuring an Organization Region Quota "region_storage_policy" block.
---

# vcfa_region_storage_policy

Provides a data source to read a Region Storage Policy in VMware Cloud Foundation Automation. These are useful
when configuring an [Organization Region Quota](/providers/vmware/vcfa/latest/docs/resources/org_region_quota) `region_storage_policy` block.

_Used by: **Provider**_

-> To retrieve Storage Classes, use the [`vcfa_storage_class`](/providers/vmware/vcfa/latest/docs/data-sources/storage_class)
data source instead

## Example Usage

```hcl
data "vcfa_region" "region" {
  name = "my-region"
}

data "vcfa_region_storage_policy" "sp" {
  region_id = data.vcfa_region.region.id
  name      = "vSAN Default Storage Policy"
}

output "policy_id" {
  value = data.vcfa_region_storage_policy.sp.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the Region Storage Policy to read
- `region_id` - (Required) The ID of the [Region](/providers/vmware/vcfa/latest/docs/data-sources/region) where the Storage Policy belongs

## Attribute reference

- `description` - Description of the Region Storage Policy
- `status` - The creation status of the Region Storage Policy. Can be `NOT_READY` or `READY`
- `storage_capacity_mib` - Storage capacity in mebibytes for this Region Storage Policy
- `storage_consumed_mib` - Consumed storage in mebibytes for this Region Storage Policy
- `kubernetes_compliant_name` - Kubernetes-compliant name of the Region Storage Policy. Requires VCFA 9.1.0 or later.
- `is_restricted` - Whether the Region Storage Policy is restricted to system-owned namespaces. Requires VCFA 9.1.2 or later.
- `storage_capacity_mb` - **Deprecated**, use `storage_capacity_mib`. The value is expressed in mebibytes
- `storage_consumed_mb` - **Deprecated**, use `storage_consumed_mib`. The value is expressed in mebibytes
