---
page_title: "VMware Cloud Foundation Automation: vcfa_nsx_manager"
subcategory: ""
description: |-
  Provides a data source to manage NSX Managers attached to VMware Cloud Foundation Automation.
---

# Resource: vcfa_nsx_manager

Provides a data source to manage NSX Managers attached to VMware Cloud Foundation Automation.

_Used by: **Provider**_

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

- `name` - (Required) A name for the NSX Manager
- `description` - (Optional) An optional description for NSX Manager
- `username` - (Required) A username for authenticating to NSX Manager
- `password` - (Required) A password for authenticating to NSX Manager
- `url` - (Required) An URL of NSX Manager
- `auto_trust_certificate` - (Required) Defines if the certificate of a given NSX Manager should
  automatically be added to trusted certificate store. **Note:** not having the certificate trusted
  will cause malfunction.

## Attribute Reference

The following attributes are exported on this resource:

- `active` - Indicates whether the NSX Manager can or cannot be used to manage networking constructs within VCFA.
- `cluster_id` - Cluster ID of the NSX Manager. Each NSX installation has a single cluster.
- `is_dedicated_for_classic_tenants` - Whether this NSX Manager is dedicated for legacy VRA-style tenants only and unable to
  participate in modern constructs such as Regions and Zones. Legacy VRA-style is deprecated and this field exists for
  the purpose of VRA backwards compatibility only
- `status` - Status of NSX Manager. One of:
  - `PENDING` - Desired entity configuration has been received by system and is pending realization.
  - `CONFIGURING` - The system is in process of realizing the entity.
  - `REALIZED` - The entity is successfully realized in the system.
  - `REALIZATION_FAILED` - There are some issues and the system is not able to realize the entity.
  - `UNKNOWN` - Current state of entity is unknown.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing NSX Manager configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

```shell
terraform import vcfa_nsx_manager.imported my-nsx-manager
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-nsx-manager` NSX Manager settings that are defined at Provider (System) level.

After that, you can expand the configuration file and either update or delete the NSX Manager as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the NSX Manager's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
