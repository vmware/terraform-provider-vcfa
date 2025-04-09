---
page_title: "VMware Cloud Foundation Automation: vcfa_region"
subcategory: ""
description: |-
  Provides a resource to manage Regions in VMware Cloud Foundation Automation.
---

# vcfa_region

Provides a resource to manage Regions in VMware Cloud Foundation Automation.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_vcenter" "one" {
  name = "vcenter-one"
}

data "vcfa_supervisor" "one" {
  name       = "my-supervisor-name"
  vcenter_id = data.vcfa_vcenter.one.id
}

data "vcfa_nsx_manager" "main" {
  name = "nsx-manager-one"
}

resource "vcfa_region" "one" {
  name                 = "region-one"
  nsx_manager_id       = data.vcfa_nsx_manager.main.id
  supervisor_ids       = [data.vcfa_supervisor.one.id]
  storage_policy_names = ["vSAN Default Storage Policy"]
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) A name for Region. It must match RFC 1123 Label name (lower-case alphabet,
  numbers between 0 and 9 and hyphen `-`)
- `description` - (Optional) An optional description for Region
- `nsx_manager_id` - (Required) NSX Manager assigned to this region. Can be looked up using
  [`vcfa_nsx_manager`](/providers/vmware/vcfa/latest/docs/data-sources/nsx_manager)
- `supervisor_ids` - (Required) A set of Supervisor IDs. At least one is required. Can be looked up
  using [`vcfa_supervisor`](/providers/vmware/vcfa/latest/docs/data-sources/supervisor)
- `storage_policy_names` - (Required) A set of Storage Policy names to be used for this region. At
  least one is required.

## Attribute Reference

The following attributes are exported on this resource:

- `cpu_capacity_mhz` - Total CPU resources in MHz available to this Region
- `cpu_reservation_capacity_mhz` - Total CPU reservation resources in MHz available to this Region
- `memory_capacity_mib` - Total memory resources (in mebibytes) available to this Region
- `memory_reservation_capacity_mib` - Total memory reservation resources (in mebibytes) available to this Region
- `status` - The creation status of the Region. Possible values are `READY`, `NOT_READY`, `ERROR`,
  `FAILED`. A Region needs to be ready and enabled to be usable

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Region configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

```
terraform import vcfa_region.imported my-region
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-region` Region settings.

After that, you must expand the configuration file before you can either update or delete the Region configuration. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources