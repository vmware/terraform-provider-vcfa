---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_region_quota_storage_policy"
sidebar_current: "docs-vcfa-resource-org-region-quota-storage-policy"
description: |-
  Provides a resource to manage Storage Policies of Organization Region Quotas in VMware Cloud Foundation Automation.
---

# vcfa\_org\_region\_quota\_storage\_policy

Provides a resource to manage Storage Policies of Organization Region Quotas in VMware Cloud Foundation Automation.

## Example Usage

```hcl
```

## Argument Reference

The following arguments are supported:

- `org_region_quota_id` - (Required) Parent [Region Quota](/providers/vmware/vcfa/latest/docs/data-sources/org_region_quota) ID
- `region_storage_policy_id` - (Required) The parent [Region Storage Policy](/providers/vmware/vcfa/latest/docs/data-sources/region_storage_policy) for this Storage Policy
- `storage_limit_mib` - (Required) Maximum allowed storage allocation in mebibytes. Minimum value: `0`

## Attribute Reference

The following attributes are exported on this resource:

- `name` - The name of the Region Quota Storage Policy. It follows RFC 1123 Label Names to conform with Kubernetes standards
- `storage_used_mib` - Amount of storage used in mebibytes

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Region Quota Storage Policy configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

TODO!!!!!!