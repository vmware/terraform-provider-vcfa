---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_vcenter"
sidebar_current: "docs-vcfa-resource-vcenter"
description: |-
  Provides a resource to manage vCenters in VMware Cloud Foundation Automation.
---

# vcfa\_vcenter

Provides a resource to manage vCenters in VMware Cloud Foundation Automation.

~> This resource can only be used by **System Administrators**

## Example Usage

```hcl
data "vcfa_nsx_manager" "demo" {
  name = "nsx-manager-one"
}

resource "vcfa_vcenter" "demo" {
  name                    = "my-vCenter"
  url                     = "https://host"
  auto_trust_certificate  = true
  refresh_vcenter_on_read = true
  username                = "admin@vsphere.local"
  password                = "CHANGE-ME"
  is_enabled              = true
  nsx_manager_id          = data.vcfa_nsx_manager.demo.id
}
```

## Argument Reference

The following arguments are supported:

* `nsx_manager_id` - (Required) ID of existing [NSX Manager][vcfa_nsx_manager-ds] that this vCenter server uses
* `name` - (Required) A name for vCenter server
* `description` - (Optional) An optional description for vCenter server
* `username` - (Required) A username for authenticating to vCenter server
* `password` - (Required) A password for authenticating to vCenter server
* `refresh_vcenter_on_create` - (Optional) An optional flag to trigger refresh operation on the
  underlying vCenter once after creation. This might take some time, but can help to load up new
  artifacts from vCenter (e.g. [Supervisors][vcfa_supervisor-ds]). This operation is visible as a new task in UI. Update
  is a no-op. It may be useful after adding vCenter or if new infrastructure is added to vCenter.
  Default `false`.
* `refresh_policies_on_create` - (Optional) An optional flag to trigger policy refresh operation on
  the underlying vCenter once after creation. This might take some time, but can help to load up new
  artifacts from vCenter (e.g. Storage Policies). Update is a no-op. This operation is visible as a
  new task in UI. It may be useful after adding vCenter or if new infrastructure is added to
  vCenter. Default `false`. 
* `refresh_vcenter_on_read` - (Optional) An optional flag to trigger refresh operation on the
  underlying vCenter on every read. This might take some time, but can help to load up new artifacts
  from vCenter (e.g. [Supervisors][vcfa_supervisor-ds]). This operation is visible as a new task in UI. Update is a no-op.
  It may be useful after adding vCenter or if new infrastructure is added to vCenter. Default
  `false`.
* `refresh_policies_on_read` - (Optional) An optional flag to trigger policy refresh operation on
  the underlying vCenter on every read. This might take some time, but can help to load up new
  artifacts from vCenter (e.g. [Storage Policies][vcfa_storage_class-ds]). Update is a no-op. This operation is visible as a
  new task in UI. It may be useful after adding vCenter or if new infrastructure is added to
  vCenter. Default `false`. 
* `url` - (Required) An URL of vCenter server
* `auto_trust_certificate` - (Required) Defines if the certificate of a given vCenter server should
  automatically be added to trusted certificate store. **Note:** not having the certificate trusted
  will cause malfunction.
* `is_enabled` - (Optional) Defines if the vCenter is enabled. Default `true`. The vCenter must
  always be disabled before removal (this resource will disable it automatically on destroy).


## Attribute Reference

The following attributes are exported on this resource:

* `has_proxy` - Indicates that this vCenter has a proxy configuration for access by authorized
  end-users
* `is_connected` - Defines if the vCenter server is connected.
* `mode` - One of `NONE`, `IAAS` (scoped to the provider), `SDDC` (scoped to tenants), `MIXED` (both
  uses are possible)
* `connection_status` - `INITIAL`, `INVALID_SETTINGS`, `UNSUPPORTED`, `DISCONNECTED`, `CONNECTING`,
  `CONNECTED_SYNCING`, `CONNECTED`, `STOP_REQ`, `STOP_AND_PURGE_REQ`, `STOP_ACK`
* `cluster_health_status` - Cluster health status. One of `GRAY` , `RED` , `YELLOW` , `GREEN`
* `version` - vCenter version
* `uuid` - UUID of vCenter
* `vcenter_host` - Host of vCenter server
* `status` - Status can be `READY` or `NOT_READY`. It is a derivative field of `is_connected` and
  `connection_status` so relying on those fields could be more precise.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing vCenter configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

```
terraform import vcfa_vcenter.imported my-vcenter
```

NOTE: the default separator (.) can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-vcenter` vCenter settings that are defined at provider level.

After that, you must expand the configuration file before you can either update or delete the vCenter configuration. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_nsx_manager-ds]: /providers/vmware/vcfa/latest/docs/data-sources/nsx_manager
[vcfa_supervisor-ds]: /providers/vmware/vcfa/latest/docs/data-sources/supervisor
[vcfa_storage_class-ds]: /providers/vmware/vcfa/latest/docs/data-sources/storage_class