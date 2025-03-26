---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_provider_gateway"
sidebar_current: "docs-vcfa-resource-provider-gateway"
description: |-
  Provides a VMware Cloud Foundation Automation Provider Gateway resource.
---

# vcfa\_provider\_gateway

Provides a VMware Cloud Foundation Automation Provider Gateway resource.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "region-one"
}

data "vcfa_tier0_gateway" "demo" {
  name      = "my-tier0-gateway"
  region_id = data.vcfa_region.demo.id
}

data "vcfa_ip_space" "demo" {
  name      = "demo-ip-space"
  region_id = data.vcfa_region.demo.id
}

data "vcfa_ip_space" "demo2" {
  name      = "demo-ip-space-2"
  region_id = data.vcfa_region.demo.id
}

resource "vcfa_provider_gateway" "demo" {
  name                  = "Demo Provider Gateway"
  description           = "Terraform Provider Gateway"
  region_id             = data.vcfa_region.demo.id
  nsxt_tier0_gateway_id = data.vcfa_tier0_gateway.demo.id
  ip_space_ids          = [data.vcfa_ip_space.demo.id, data.vcfa_ip_space.demo2.id]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for Provider Gateway
* `description` - (Optional) A description for Provider Gateway
* `region_id` - (Required) A Region ID for Provider Gateway
* `nsxt_tier0_gateway_id` - (Required) An existing NSX-T Tier 0 Gateway
* `ip_space_ids` - (Required) A set of IP Space IDs that should be assigned to this Provider Gateway

## Attribute Reference

* `status` - Current status of the entity. Possible values are:
 * `PENDING` - Desired entity configuration has been received by system and is pending realization
 * `CONFIGURING` - The system is in process of realizing the entity
 * `REALIZED` - The entity is successfully realized in the system
 * `REALIZATION_FAILED` - There are some issues and the system is not able to realize the entity
 * `UNKNOWN` - Current state of entity is unknown

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Provider Gateway configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

```
terraform import vcfa_provider_gateway.imported my-region-name.my-provider-gateway
```

NOTE: the default separator (.) can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-provider-gateway` Provider Gateway in Region `my-region-name`

After that, you can expand the configuration file and either update or delete the Provider Gateway as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Provider Gateway's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources