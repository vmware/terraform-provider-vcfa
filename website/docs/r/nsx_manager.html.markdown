---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_nsx_manager"
sidebar_current: "docs-vcfa-resource-nsx-manager"
description: |-
  Provides a data source to manage NSX Managers attached to VMware Cloud Foundation Automation.
---

# vcfa\_nsx\_manager

Provides a data source to manage NSX Managers attached to VMware Cloud Foundation Automation.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
resource "vcfa_nsx_manager" "test" {
  name                   = "nsx-manager-one"
  description            = "terraform test"
  username               = "admin"
  password               = "CHANGE-ME"
  url                    = "https://HOST"
  auto_trust_certificate = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for NSX Manager
* `description` - (Optional) An optional description for NSX Manager
* `username` - (Required) A username for authenticating to NSX Manager
* `password` - (Required) A password for authenticating to NSX Manager
* `url` - (Required) An URL of NSX Manager
* `auto_trust_certificate` - (Required) Defines if the certificate of a given NSX Manager should
  automatically be added to trusted certificate store. **Note:** not having the certificate trusted
  will cause malfunction.
* `network_provider_scope` - (Optional) The network provider scope is the tenant facing name for the
  NSX Manager.

## Attribute Reference

The following attributes are exported on this resource:

* `status` - Status of NSX Manager. One of:
* `PENDING` - Desired entity configuration has been received by system and is pending realization.
* `CONFIGURING` - The system is in process of realizing the entity.
* `REALIZED` - The entity is successfully realized in the system.
* `REALIZATION_FAILED` - There are some issues and the system is not able to realize the entity.
* `UNKNOWN` - Current state of entity is unknown.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing NSX Manager configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcfa_nsx_manager.imported my-nsx-manager
```

The above would import the `my-nsx-manager` NSX Manager settings that are defined at provider
level.