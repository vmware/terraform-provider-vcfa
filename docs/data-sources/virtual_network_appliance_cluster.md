---
page_title: "VMware Cloud Foundation Automation: vcfa_virtual_network_appliance_cluster"
subcategory: ""
description: |-
  Provides a data source to read a Virtual Network Appliance Cluster in VMware Cloud Foundation Automation.
---

# vcfa_virtual_network_appliance_cluster

Provides a data source to read a Virtual Network Appliance (VNA) Cluster in VMware Cloud Foundation Automation. VNA
Clusters are created in NSX and imported into VCFA with a synchronization. It is useful when configuring
[`vcfa_org_regional_networking`](/providers/vmware/vcfa/latest/docs/resources/org_regional_networking).

-> Requires VCFA 9.1.0 or later. The `service_type` attribute requires VCFA 9.1.1 or later.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "region-one"
}

data "vcfa_virtual_network_appliance_cluster" "demo" {
  name             = "my-vna-cluster"
  region_id        = data.vcfa_region.demo.id
  sync_before_read = true
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the Virtual Network Appliance Cluster
- `region_id` - (Optional) The ID of the parent Region. If set, it must match the Region of the Virtual Network Appliance
  Cluster. Can be looked up using [`vcfa_region`](/providers/vmware/vcfa/latest/docs/data-sources/region) data source
- `sync_before_read` - (Optional) Set to `true` to trigger a synchronization with NSX before attempting to search for the
  Virtual Network Appliance Cluster. Default `false`

## Attribute Reference

- `description` - Description of the Virtual Network Appliance Cluster
- `node_count` - Number of Virtual Network Appliances in the cluster. If this information is not available, it will be
  set to `-1`
- `org_count` - Number of Organizations using the cluster
- `health_status` - Current health status of the cluster. One of `UP`, `DOWN`, `DEGRADED` or `UNKNOWN`
- `service_type` - Service type of the cluster. One of `VPC_SERVICES` (used to connect VPC workloads) or
  `ROUTE_CONTROLLER` (exclusively for the route controller). Requires VCFA 9.1.1 or later
