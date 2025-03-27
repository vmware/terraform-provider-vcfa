---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org"
sidebar_current: "docs-vcfa-resource-org"
description: |-
  Provides a resource to manage VMware Cloud Foundation Automation Organizations.
---

# vcfa\_org

Provides a resource to manage VMware Cloud Foundation Automation Organizations.

_Used by: **Provider**_

## Example Usage

```hcl
resource "vcfa_org" "test" {
  name         = "terraform-org"
  display_name = "Terraform Organization"
  description  = "Created with Terraform"
  is_enabled   = true
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) A name for Organization with which users log in to it as it will be used in
  the URL. The Org must be disabled to or transition from previous disabled state
  (`is_enabled=false`) to change a name because it changes tenant login URL
- `display_name` - (Required) A human-readable name for Organization
- `description` - (Optional) An optional description for Organization
- `is_enabled` - (Optional) Defines if Organization is enabled. Default `true`. **Note:**
  Organization has to be disabled before removal and this resource will automatically disable it if
  the resource is destroyed.
- `is_classic_tenant` - (Optional) Defines if this Organization is a classic VRA style tenant. Defaults to `false`. Cannot be
  changed after creation (changing it will force the re-creation of the Organization)

## Attribute Reference

The following attributes are exported on this resource:

- `managed_by_id` - ID of Org that owns this Org
- `managed_by_name` - Name of Org that owns this Org
- `org_region_quota_count` - Number of Region Quotas belonging to this Organization
- `catalog_count` - Number of catalogs belonging to this Organization
- `vapp_count` - Number of vApps belonging to this Organization
- `running_vm_count` - Number of running VMs belonging to this Organization
- `user_count` - Number of users belonging to this Organization
- `disk_count` - Number of disks belonging to this Organization
- `can_publish` - Defines if this Organization can publish catalogs externally

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Org configuration can be [imported][docs-import] into this resource via supplying path
for it. An example is below:

```
terraform import vcfa_org.imported my-org-name
```

NOTE: the default separator (.) can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-org-name` Organization settings.

After that, you can expand the configuration file and either update or delete the Organization Settings as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Organization Settings's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources