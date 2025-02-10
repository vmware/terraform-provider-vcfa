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
data "vcfa_vcenter" "vc" {
  name = "my-vcenter"
}

data "vcfa_supervisor" "supervisor" {
  name       = "my-supervisor"
  vcenter_id = vcfa_vcenter.vc.id
}

data "vcfa_region" "one" {
  name = "region-one"
}

data "vcfa_region_zone" "one" {
  region_id = data.vcfa_region.one.id
  name      = "my-zone"
}

data "vcfa_region_vm_class" "vm_class1" {
  name      = "best-effort-4xlarge"
  region_id = data.vcfa_region.region1.id
}

data "vcfa_region_vm_class" "vm_class2" {
  name      = "best-effort-8xlarge"
  region_id = data.vcfa_region.region1.id
}

resource "vcfa_org_region_quota" "first" {
  org_id         = vcfa_org.test.id
  region_id      = data.vcfa_region.one.id
  supervisor_ids = [data.vcfa_supervisor.test.id]
  zone_resource_allocations {
    region_zone_id         = data.vcfa_region_zone.one.id
    cpu_limit_mhz          = 2000
    cpu_reservation_mhz    = 100
    memory_limit_mib       = 1024
    memory_reservation_mib = 512
  }
  region_vm_class_ids = [
    data.vcfa_region_vm_class.vm_class1.id,
    data.vcfa_region_vm_class.vm_class2.id,
  ]
}

data "vcfa_region_storage_policy" "region-sp" {
  name      = "wcplocal_storage_profile"
  region_id = data.vcfa_org_region_quota.test.region_id
}

resource "vcfa_org_region_quota_storage_policy" "rq-sp" {
  org_region_quota_id      = vcfa_org_region_quota.first.id
  region_storage_policy_id = data.vcfa_region_storage_policy.region-sp.id
  storage_limit_mib        = 100
}
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