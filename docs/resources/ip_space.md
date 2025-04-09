---
page_title: "VMware Cloud Foundation Automation: vcfa_ip_space"
subcategory: ""
description: |-
  Provides a resource to manage IP Spaces in VMware Cloud Foundation Automation. IP spaces offer a structured approach for
  administrators to allocate IP addresses to different Organizations, enabling connectivity to external networks.
---

# vcfa_ip_space

Provides a resource to manage IP Spaces in VMware Cloud Foundation Automation. IP spaces offer a structured approach for
administrators to allocate IP addresses to different [Organizations][vcfa_org], enabling connectivity to external networks.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "demo-region"
}

resource "vcfa_ip_space" "demo" {
  name                          = "demo-ip-space"
  description                   = "description"
  region_id                     = data.vcfa_region.demo.id
  external_scope                = "12.12.0.0/30"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1

  internal_scope {
    name = "scope1"
    cidr = "10.0.0.0/24"
  }

  internal_scope {
    name = "scope2"
    cidr = "20.0.0.0/24"
  }

  internal_scope {
    name = "scope3"
    cidr = "30.0.0.0/24"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) A tenant facing name for the IP Space
- `description` - (Optional) An optional description
- `region_id` - (Required) A Region ID. Can be looked up using [Region data source][vcfa_region-ds]
- `external_scope` - (Required) A CIDR (e.g. 10.0.0.0/8) for External Reachability. It represents
  the IPs used outside the datacenter, north of the [Provider Gateway][vcfa_provider_gateway]
- `default_quota_max_subnet_size` - (Required) Maximum subnet size that can be allocated (e.g. 24)
- `default_quota_max_cidr_count` - (Required) Maximum number of CIDRs that can be allocated (`-1` for unlimited)
- `default_quota_max_ip_count` - (Required) Maximum number of floating IPs that can be allocated (`-1` for unlimited)
- `internal_scope` - (Required) A set of IP Blocks that represent IPs used in this local datacenter,
  south of the [Provider Gateway][vcfa_provider_gateway]. IPs within this scope are used for configuring services and
  networks. See [internal_scope](#internal-scope) for more details.

<a id="internal-scope"></a>

## internal_scope block

- `cidr` - (Required) CIDR for IP block (e.g. 10.0.0.0/16)
- `name` - (Optional) An optional friendly name for this block

## Attribute Reference

The following attributes are exported on this resource:

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

An existing IP Space configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

```
terraform import vcfa_ip_space.imported my-region-name.my-ip-space-name
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-ip-space-name` IP Space that is assigned to `my-region-name` [Region][vcfa_region-ds].

After that, you can expand the configuration file and either update or delete the IP Space as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the IP Space's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_region-ds]: /providers/vmware/vcfa/latest/docs/data-sources/region
[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org
[vcfa_provider_gateway]: /providers/vmware/vcfa/latest/docs/resources/provider_gateway
