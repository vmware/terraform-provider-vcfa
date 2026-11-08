---
page_title: "VMware Cloud Foundation Automation: vcfa_org_regional_networking"
subcategory: ""
description: |-
  Provides a resource to manage Organization Regional Networking Settings in VMware Cloud Foundation Automation.
---

# vcfa_org_regional_networking

Provides a resource to manage Organization Regional Networking Settings in VMware Cloud Foundation Automation.

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

- `name` - (Required) A name for Organization Regional Networking Configuration
- `org_id` - (Required) An [Organization][vcfa_org] ID for which the Regional Networking Settings are to be
  configured
- `provider_gateway_id` - (Optional) [Provider Gateway][vcfa_provider_gateway] ID that should be used for this Organization.
  Exactly one of `provider_gateway_id`, `distributed_vlan_connection_id` or `distributed_vxlan_connection_id` must be set
- `distributed_vlan_connection_id` - (Optional) [Distributed VLAN Connection][vcfa_distributed_vlan_connection-ds] ID used
  for egress, as an alternative to a Provider Gateway. Requires VCFA 9.1.0 or later
- `distributed_vxlan_connection_id` - (Optional) Distributed VXLAN Connection ID used for egress, as an alternative to a
  Provider Gateway. Requires VCFA 9.1.1 or later
- `virtual_network_appliance_cluster_id` - (Optional) Virtual Network Appliance Cluster ID used with a distributed
  connection. Requires VCFA 9.1.0 or later
- `default_vpc_private_subnet_cidr_override` - (Optional) IPv4 CIDR override for the default VPC's private subnet. It can
  only be set on creation. Requires VCFA 9.1.0 or later
- `project_tgw_private_subnet_cidr_override` - (Optional) IPv4 CIDR override for the project transit gateway's private
  subnet. It can only be set on creation. Requires VCFA 9.1.0 or later
- `region_id` - (Required) [Region][vcfa_region] ID that should be used for this Organization
- `edge_cluster_id` - (Optional) [Edge Cluster][vcfa_edge_cluster-ds] ID that can be used for this Organization. Can be left out so
  that it is picked automatically

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Regional Networking Configuration can be [imported][docs-import] into this resource via supplying path
for it. An example is below:

```shell
terraform import vcfa_org_regional_networking.imported my-org-name.my-regional-configuration-name
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-regional-configuration-name` Regional Networking Configuration Settings that are defined for `my-org-name` Organization.

After that, you can expand the configuration file and either update or delete the Regional Networking Configuration Settings as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Regional Networking Configuration Settings' stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org
[vcfa_provider_gateway]: /providers/vmware/vcfa/latest/docs/resources/provider_gateway
[vcfa_region]: /providers/vmware/vcfa/latest/docs/resources/region
[vcfa_edge_cluster-ds]: /providers/vmware/vcfa/latest/docs/data-sources/edge_cluster
[vcfa_distributed_vlan_connection-ds]: /providers/vmware/vcfa/latest/docs/data-sources/distributed_vlan_connection
