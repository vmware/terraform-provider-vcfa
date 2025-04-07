---
page_title: "VMware Cloud Foundation Automation: vcfa_edge_cluster_qos"
subcategory: ""
description: |-
  Provides a resource to define Quality of Service (QoS) settings of an Edge Cluster in VMware Cloud Foundation Automation.
---

# vcfa\_edge\_cluster\_qos

Provides a resource to define Quality of Service (QoS) settings of an Edge Cluster in VMware Cloud Foundation Automation.

_Used by: **Provider**_

-> This resource does not create an Edge Cluster QoS entity, but configures QoS for a given
`edge_cluster_id`. Similarly, `terraform destroy` operation does not remove Edge Cluster, but resets
QoS settings to default (unlimited).

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "region-one"
}

data "vcfa_edge_cluster" "demo" {
  name             = "my-edge-cluster"
  region_id        = data.vcfa_region.demo.id
  sync_before_read = true
}

resource "vcfa_edge_cluster_qos" "demo" {
  edge_cluster_id = data.vcfa_edge_cluster.demo.id

  egress_committed_bandwidth_mbps  = 1
  egress_burst_size_bytes          = 2
  ingress_committed_bandwidth_mbps = 3
  ingress_burst_size_bytes         = 4
}
```

## Argument Reference

The following arguments are supported:

- `edge_cluster_id` - (Required) An ID of Edge Cluster. Can be looked up using
  [vcfa_edge_cluster](/providers/vmware/vcfa/latest/docs/data-sources/edge_cluster) data source
- `egress_committed_bandwidth_mbps` - (Optional) Committed egress bandwidth specified in Mbps.
  Bandwidth is limited to line rate. Traffic exceeding bandwidth will be dropped. Required with
  `egress_burst_size_bytes`. Default is `-1` (unlimited)
- `egress_burst_size_bytes` - (Optional) Egress burst size in bytes. Required with
  `egress_committed_bandwidth_mbps`. Default is `-1` (unlimited)
- `ingress_committed_bandwidth_mbps` - (Optional) Committed ingress bandwidth specified in Mbps.
  Bandwidth is limited to line rate. Traffic exceeding bandwidth will be dropped. Required with
  `ingress_burst_size_bytes`. Default is `-1` (unlimited)
- `ingress_burst_size_bytes` - (Optional) Ingress burst size in bytes. Required with
  `ingress_committed_bandwidth_mbps`. Default is `-1` (unlimited)

-> Deleting this resource will reset all values to unlimited

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Edge Cluster QoS configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

```
terraform import vcfa_edge_cluster_qos.imported my-region-name.my-edge-cluster-name
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-edge-cluster-name` Edge Cluster QoS settings that is in `my-region-name` Region.

After that, you can expand the configuration file and either update or delete the Edge Cluster QoS settings as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources