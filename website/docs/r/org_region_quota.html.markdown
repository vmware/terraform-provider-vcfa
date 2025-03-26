---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_region_quota"
sidebar_current: "docs-vcfa-resource-org-region-quota"
description: |-
  Provides a resource to manage VMware Cloud Foundation Automation Organization Region Quotas.
---

# vcfa\_org\_region\_quota

Provides a resource to manage VMware Cloud Foundation Automation Organization Region Quotas.

~> This resource can only be used by **System Administrators**

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

data "vcfa_region_storage_policy" "region-sp" {
  name      = "vSAN Default Storage Policy"
  region_id = vcfa_region.region1.id
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
  region_storage_policy {
    region_storage_policy_id = data.vcfa_region_storage_policy.region-sp.id
    storage_limit_mib        = 1024
  }
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) An Org ID for this Organization Region Quota to be assigned to
- `region_id` - (Required) A Region ID that this Organization Region Quota should be backed by
- `supervisor_ids` - (Required) A set of Supervisor IDs that back this Organization Region Quota. Can be looked up
  using [`vcfa_supervisor`](/providers/vmware/vcfa/latest/docs/data-sources/supervisor) data source
- `zone_resource_allocations` - (Required) A set of Zone Resource Allocation definitions. See [Zone Resource Allocations](#zone-resource-allocations-block)
- `region_vm_class_ids` - (Required) A set of Region VM Class IDs. These can be fetched with [`vcfa_region_vm_class` data source](/providers/vmware/vcfa/latest/docs/data-sources/region_vm_class)
- `region_storage_policy` - (Required) A set of Region Storage Policies. See [Region Storage Policies](#region-storage-policies)

<a id="zone-resource-allocations-block"></a>
## Zone Resource Allocations

- `region_zone_id` - (Required) Can be looked up using
  [`vcfa_region_zone`](/providers/vmware/vcfa/latest/docs/data-sources/region_zone) data source
- `cpu_limit_mhz` - (Required) Maximum CPU consumption limit in MHz
- `cpu_reservation_mhz` - (Required) Defines reserved CPU capacity in MHz
- `memory_limit_mib` - (Required) Maximum memory consumption limit in MiB
- `memory_reservation_mib` - (Required) Defines reserved memory capacity in MiB

A computed attribute `region_zone_name` will be set in each `zone_resource_allocations` block.

<a id="region-storage-policies"></a>
## Region Storage Policies

- `region_storage_policy_id` - The ID of a Region Storage Policy. It can be fetched with [`vcfa_region_storage_policy` data source](/providers/vmware/vcfa/latest/docs/data-sources/region_storage_policy).
- `storage_limit_mib` - Maximum allowed storage allocation in mebibytes. Minimum value: `0`

Each block defines some read-only attributes:

- `id` - ID of the Region Quota Storage Policy
- `name` - The name of the Region Quota Storage Policy. It follows RFC 1123 Label Names to conform with Kubernetes standards
- `storage_used_mib` - Amount of storage used in mebibytes

## Attribute Reference

The following attributes are exported on this resource:

- `name` - The name of the Organization Region Quota, it's assigned on creation and can't be changed
- `status` - The creation status of the Organization Region Quota. Possible values are `READY`, `NOT_READY`, `ERROR`,
  `FAILED`

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Organization Region Quota configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is
below:

```
terraform import vcfa_org_region_quota.imported my-org-name.my-region-name
```

NOTE: the default separator (.) can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the Organization Region Quota that belongs to `my-org-name` Organization and `my-region-name` Region.

After that, you can expand the configuration file and either update or delete the Organization Region Quota as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Organization Region Quota's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources