---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_regional_networking"
sidebar_current: "docs-vcfa-resource-org-regional-networking"
description: |-
  Provides a resource to manage VMware Cloud Foundation Automation Organization Regional Networking Settings.
---

# vcfa\_org\_regional\_networking

Provides a resource to manage VMware Cloud Foundation Automation Organization Regional Networking Settings.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "my-org-name"
}

data "vcfa_region" "demo" {
  name = "region-one"
}

data "vcfa_provider_gateway" "demo" {
  name      = "provider-gateway"
  region_id = vcfa_region.region.id
}

data "vcfa_edge_cluster" "demo" {
  name      = "edge-cluster-1"
  region_id = data.vcfa_region.demo.id
}

resource "vcfa_org_networking" "demo" {
  org_id   = data.vcfa_org.demo.id
  log_name = "org1"
}

resource "vcfa_org_regional_networking" "demo" {
  name = "net-one"

  # log_name in vcfa_org_networking must be set before therefore using 
  # vcfa_org_regional_networking.demo.id that also contains Org ID
  # to make correct order of actions
  org_id = vcfa_org_networking.demo.id

  provider_gateway_id = data.vcfa_provider_gateway.demo.id
  region_id           = data.vcfa_region.demo.id

  edge_cluster_id = data.vcfa_edge_cluster.test.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) A name for Org Regional Networking Configuration
- `org_id` - (Required) An Organization ID for which the Regional Networking Settings are to be
  configured
- `provider_gateway_id` - (Required) Provider Gateway ID that should be used for this Org
- `region_id` - (Required) Region ID that should be used for this Org
- `edge_cluster_id` - (Optional) Edge Cluster ID that can be used for this Org. Can be left out so
  that it is picked automatically

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Regional Networking Configuration can be [imported][docs-import] into this resource via supplying path
for it. An example is below:

```
terraform import vcfa_org_regional_networking.imported my-org-name.my-regional-configuration-name
```

NOTE: the default separator (.) can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-regional-configuration-name` Regional Networking Configuration Settings that are defined for `my-org-name` Organization.

After that, you can expand the configuration file and either update or delete the Regional Networking Configuration Settings as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Regional Networking Configuration Settings' stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources