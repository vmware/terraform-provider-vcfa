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
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1
  provider_visibility_only      = true

  cidr_blocks {
    name = "block1"
    cidr = "10.0.0.0/24"
  }

  cidr_blocks {
    name = "block2"
    cidr = "20.0.0.0/24"
  }

  ip_address_ranges {
    start_ip_address = "10.0.1.1"
    end_ip_address   = "10.0.1.255"
  }

  reserved_ip_address_ranges {
    start_ip_address = "10.0.0.1"
    end_ip_address   = "10.0.0.10"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) A tenant facing name for the IP Space
- `description` - (Optional) An optional description
- `region_id` - (Required) A Region ID. Can be looked up using [Region data source][vcfa_region-ds]
- `external_scope` - (Optional, **Deprecated**) A CIDR for External Reachability. Use
  `inbound_remote_networks` in [`vcfa_provider_gateway`][vcfa_provider_gateway] instead
- `default_quota_max_subnet_size` - (Required) Maximum subnet size that can be allocated (e.g. 24)
- `default_quota_max_cidr_count` - (Required) Maximum number of CIDRs that can be allocated (`-1` for unlimited)
- `default_quota_max_ip_count` - (Required) Maximum number of floating IPs that can be allocated (`-1` for unlimited)
- `cidr_blocks` - (Optional) A set of CIDR blocks. Along with `ip_address_ranges`, typically defines the span of IP
  addresses used within a Data Center. At least one of `cidr_blocks`, `internal_scope` or `ip_address_ranges` is
  required. Conflicts with `internal_scope`. See [cidr_blocks](#cidr_blocks-block) for more details
- `internal_scope` - (Optional, **Deprecated**) Use `cidr_blocks` instead. A set of CIDR blocks. Along with
  `ip_address_ranges`, typically defines the span of IP addresses used within a Data Center. At least one of
  `cidr_blocks`, `internal_scope` or `ip_address_ranges` is required. Conflicts with `cidr_blocks`. See
  [cidr_blocks](#cidr_blocks-block) for the block structure
- `ip_address_ranges` - (Optional) A set of IP address ranges. Along with `cidr_blocks`, typically defines the
  span of IP addresses used within a Data Center. At least one of `cidr_blocks`, `internal_scope` or
  `ip_address_ranges` is required. See [ip_address_ranges](#ip_address_ranges-block) for more details
- `provider_visibility_only` - (Optional) If set to `true`, the IP Space details will be hidden from organizations
- `reserved_ip_address_ranges` - (Optional) IP addresses that will not be considered for IP allocation. Reserved
  IPs have to be part of one of the CIDRs or IP Ranges. See [ip_address_ranges](#ip_address_ranges-block) for the
  block structure

## cidr_blocks block

- `cidr` - (Required) CIDR for IP block (e.g. `10.0.0.0/16`)
- `name` - (Optional) An optional friendly name for this block

## ip_address_ranges block

- `start_ip_address` - (Required) Starting IP address in the range
- `end_ip_address` - (Required) Ending IP address in the range

## Attribute Reference

The following attributes are exported on this resource:

- `backing_id` - ID for the matching IP Block in NSX
- `is_imported_ip_block` - Indicates if the IP Block is imported from an existing NSX IP Block
- `subnet_exclusive` - Whether this IP Block is exclusively for a single CIDR
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

```shell
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
