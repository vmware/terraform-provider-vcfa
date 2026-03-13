---
page_title: "VMware Cloud Foundation Automation: vcfa_shared_subnet"
subcategory: ""
description: |-
  Provides a resource to manage Shared Subnets in VMware Cloud Foundation Automation. Shared Subnets define Layer 2 network segments within a Region that can be shared across Organizations.
---

# vcfa_shared_subnet

Provides a resource to manage Shared Subnets in VMware Cloud Foundation Automation. Shared Subnets define Layer 2 network segments within a [Region][vcfa_region] that can be shared across [Organizations][vcfa_org].

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "demo-region"
}

resource "vcfa_shared_subnet" "demo" {
  name         = "demo-shared-subnet"
  description  = "My Shared Subnet"
  region_id    = data.vcfa_region.demo.id
  subnet_type  = "VLAN"
  gateway_cidr = "10.0.0.1/24"
  vlan_id      = 100
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) A name for the Shared Subnet
- `description` - (Optional) A description for the Shared Subnet
- `region_id` - (Required) The ID of the [Region][vcfa_region] for this Shared Subnet.
  This field cannot be updated after creation. Can be looked up using the [Region data source][vcfa_region-ds]
- `subnet_type` - (Required) The type of the Shared Subnet (e.g. `VLAN`). This field cannot be updated after creation
- `gateway_cidr` - (Required) The gateway CIDR for the Shared Subnet (e.g. `10.0.0.1/24`). This field cannot be updated after creation
- `vlan_id` - (Required) The VLAN ID for the Shared Subnet

## Attribute Reference

The following attributes are exported on this resource:

- `backing_id` - ID for the matching Subnet in NSX
- `ip_space_id` - The ID of the [IP Space][vcfa_ip_space] that is automatically created for this Shared Subnet
- `status` - One of:
  - `PENDING` - Desired entity configuration has been received by system and is pending realization
  - `CONFIGURING` - The system is in process of realizing the entity
  - `REALIZED` - The entity is successfully realized in the system
  - `REALIZATION_FAILED` - There are some issues and the system is not able to realize the entity
  - `UNKNOWN` - Current state of entity is unknown

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Shared Subnet configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

```shell
terraform import vcfa_shared_subnet.imported my-region-name.my-shared-subnet-name
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-shared-subnet-name` Shared Subnet in Region `my-region-name`.

After that, you can expand the configuration file and either update or delete the Shared Subnet as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Shared Subnet's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_region]: /providers/vmware/vcfa/latest/docs/resources/region
[vcfa_region-ds]: /providers/vmware/vcfa/latest/docs/data-sources/region
[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org
[vcfa_ip_space]: /providers/vmware/vcfa/latest/docs/resources/ip_space
