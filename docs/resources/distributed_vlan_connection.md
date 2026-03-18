---
page_title: "VMware Cloud Foundation Automation: vcfa_distributed_vlan_connection"
subcategory: ""
description: |-
  Provides a resource to manage Distributed VLAN Connections in VMware Cloud Foundation Automation. Distributed VLAN Connections define external network connectivity over VLAN-backed segments within a Region.
---

# vcfa_distributed_vlan_connection

Provides a resource to manage Distributed VLAN Connections in VMware Cloud Foundation Automation. Distributed VLAN Connections define external network connectivity over VLAN-backed segments within a [Region][vcfa_region].

_Used by: **Provider**_

## Example Usage

### With an existing IP Space

```hcl
data "vcfa_region" "demo" {
  name = "demo-region"
}

data "vcfa_ip_space" "demo" {
  name      = "demo-ip-space"
  region_id = data.vcfa_region.demo.id
}

resource "vcfa_distributed_vlan_connection" "demo" {
  name             = "demo-distributed-vlan-connection"
  description      = "My Distributed VLAN Connection"
  region_id        = data.vcfa_region.demo.id
  gateway_cidr     = "32.0.1.1/24"
  ip_space_id      = data.vcfa_ip_space.demo.id
  subnet_exclusive = false
  vlan_id          = 100
}
```

### With auto-created IP Space

```hcl
data "vcfa_region" "demo" {
  name = "demo-region"
}

resource "vcfa_distributed_vlan_connection" "demo" {
  name             = "demo-distributed-vlan-connection"
  description      = "My Distributed VLAN Connection"
  region_id        = data.vcfa_region.demo.id
  gateway_cidr     = "10.0.0.1/24"
  subnet_exclusive = true
  vlan_id          = 110
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) A name for the Distributed VLAN Connection
- `description` - (Optional) A description for the Distributed VLAN Connection
- `region_id` - (Required) The ID of the [Region][vcfa_region] for this Distributed VLAN Connection.
  This field cannot be updated after creation. Can be looked up using the [Region data source][vcfa_region-ds]
- `gateway_cidr` - (Required) The gateway CIDR for the Distributed VLAN Connection (e.g. `32.0.1.1/24`)
- `ip_space_id` - (Optional) Reference to an [IP Space][vcfa_ip_space] that is used for the external connection.
  Required when `subnet_exclusive` is `false`. When `subnet_exclusive` is `true` and this is not provided, an
  IP Space is automatically created
- `subnet_exclusive` - (Required) Whether this Distributed VLAN Connection is exclusively for the gateway CIDR only. This field cannot be updated after creation
- `vlan_id` - (Required) The VLAN ID for the external traffic
- `zone_ids` - (Optional) A set of Supervisor Zone IDs that this Distributed VLAN Connection spans

## Attribute Reference

The following attributes are exported on this resource:

- `backing_id` - ID for the matching Distributed VLAN Connection in NSX
- `ip_space_id` - The ID of the [IP Space][vcfa_ip_space] used for the external connection (also available when auto-created)
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

An existing Distributed VLAN Connection configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

```shell
terraform import vcfa_distributed_vlan_connection.imported my-region-name.my-distributed-vlan-connection-name
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-distributed-vlan-connection-name` Distributed VLAN Connection in Region `my-region-name`.

After that, you can expand the configuration file and either update or delete the Distributed VLAN Connection as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Distributed VLAN Connection's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_region]: /providers/vmware/vcfa/latest/docs/resources/region
[vcfa_region-ds]: /providers/vmware/vcfa/latest/docs/data-sources/region
[vcfa_ip_space]: /providers/vmware/vcfa/latest/docs/resources/ip_space
