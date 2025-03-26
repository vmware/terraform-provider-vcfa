---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_regional_networking_vpc_qos"
sidebar_current: "docs-vcfa-resource-org-regional-networking-vpc-qos"
description: |-
  Provides a resource to manage VMware Cloud Foundation Automation Organization Regional Networking VPC QoS settings.
---

# vcfa\_org\_regional\_networking\_vpc\_qos

Provides a resource to manage VMware Cloud Foundation Automation Organization Regional Networking VPC QoS settings.

-> Organization Regional Networking VPC inherits QoS settings from the Edge Cluster by default, but
one can use this resource to override provided defaults. Deleting this resource will revert the
settings back to Edge Cluster defaults.

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "my-org-name"
}

data "vcfa_org_regional_networking" "demo" {
  name   = "my-name"
  org_id = vcfa_org.demo.id
}

resource "vcfa_org_regional_networking_vpc_qos" "demo" {
  org_regional_networking_id       = vcfa_org_regional_networking.demo.id
  ingress_committed_bandwidth_mbps = 14
  ingress_burst_size_bytes         = 15
  egress_committed_bandwidth_mbps  = 16
  egress_burst_size_bytes          = 17
}
```

## Argument Reference

The following arguments are supported:

- `org_regional_networking_id` - (Required) The ID of Organization Regional Networking configuration
- `egress_committed_bandwidth_mbps` - (Optional) Committed egress bandwidth specified in Mbps.
  Bandwidth is limited to line rate. Traffic exceeding bandwidth will be dropped. Required with
  `egress_burst_size_bytes`. Default is `-1` - unlimited
- `egress_burst_size_bytes` - (Optional) Egress burst size in bytes. Required with
  `egress_committed_bandwidth_mbps`. Default is `-1` - unlimited
- `ingress_committed_bandwidth_mbps` - (Optional) Committed ingress bandwidth specified in Mbps.
  Bandwidth is limited to line rate. Traffic exceeding bandwidth will be dropped. Required with
  `ingress_burst_size_bytes`. Default is `-1` - unlimited
- `ingress_burst_size_bytes` - (Optional) Ingress burst size in bytes. Required with
  `ingress_committed_bandwidth_mbps`. Default is `-1` - unlimited

## Attribute Reference

- `edge_cluster_id` - ID of parent Edge Cluster

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Regional Networking Configuration can be [imported][docs-import] into this resource via supplying path
for it. An example is below:

```
terraform import vcfa_org_regional_networking_vpc_qos.imported my-org-name.my-regional-configuration-name
```

NOTE: the default separator (.) can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the Regional Networking VPC QoS settings that are defined by the `my-regional-configuration-name`
Regional Networking Configuration Settings present in the `my-org-name` Organization.

After that, you can expand the configuration file and either update or delete the Regional Networking VPC QoS as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Regional Networking VPC QoS' stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources