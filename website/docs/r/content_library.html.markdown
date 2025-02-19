---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_content_library"
sidebar_current: "docs-vcfa-resource-content-library"
description: |-
  Provides a VMware Cloud Foundation Automation Content Library resource. This can be used to manage Content Libraries.
---

# vcfa\_content\_library

Provides a VMware Cloud Foundation Automation Content Library resource. This can be used to manage Content Libraries.

## Example Usage for a Provider Content Library

```hcl
data "vcfa_region" "region" {
  name = "My Region"
}

data "vcfa_storage_class" "sc" {
  region_id = data.vcfa_region.region.id
  name      = "vSAN Default Storage Policy"
}

resource "vcfa_content_library" "cl" {
  name        = "My Library"
  description = "A simple library"
  storage_class_ids = [
    data.vcfa_storage_class.sc.id
  ]
}
```

## Example Usage for a Tenant Content Library as an Administrator

```hcl
data "vcfa_vcenter" "vc" {
  name = "my-vcenter"
}

data "vcfa_nsx_manager" "nsx_manager" {
  name = "my-nsx-manager"
}

data "vcfa_region" "region" {
  name = "my-region"
}

data "vcfa_region_vm_class" "region_vm_class0" {
  region_id = data.vcfa_region.region.id
  name      = "best-effort-2xlarge"
}

resource "vcfa_org" "test" {
  name         = "my-org"
  display_name = "my-org"
  description  = "my-org"
}

data "vcfa_supervisor" "test" {
  name       = "supervisor"
  vcenter_id = data.vcfa_vcenter.vc.id
  depends_on = [data.vcfa_vcenter.vc]
}

data "vcfa_region_zone" "test" {
  region_id = data.vcfa_region.region.id
  name      = "region-zone"
}

data "vcfa_region_storage_policy" "sp" {
  name      = "default_storage_policy"
  region_id = data.vcfa_region.region.id
}

resource "vcfa_org_region_quota" "test" {
  org_id         = vcfa_org.test.id
  region_id      = data.vcfa_region.region.id
  supervisor_ids = [data.vcfa_supervisor.test.id]
  zone_resource_allocations {
    region_zone_id         = data.vcfa_region_zone.test.id
    cpu_limit_mhz          = 1900
    cpu_reservation_mhz    = 90
    memory_limit_mib       = 500
    memory_reservation_mib = 200
  }
  region_vm_class_ids = [
    data.vcfa_region_vm_class.region_vm_class0.id,
    data.vcfa_region_vm_class.region_vm_class1.id
  ]
  region_storage_policy {
    region_storage_policy_id = data.vcfa_region_storage_policy.sp.id
    storage_limit_mib        = 1024
  }
}

data "vcfa_storage_class" "sc" {
  region_id = data.vcfa_region.region.id
  name      = data.vcfa_region_storage_policy.sp.name
}

resource "vcfa_content_library" "cl1" {
  org_id      = vcfa_org.test.id
  name        = "my-content-library"
  description = "Example ibrary"
  auto_attach = false # Defaults to true
  storage_class_ids = [
    data.vcfa_storage_class.sc.id
  ]
  delete_recursive = true

  # We need an available Region Quota
  depends_on = [ vcfa_org_region_quota.test]
}
```

## Example Usage for a Tenant Content Library as a Tenant User

```hcl
data "vcfa_region" "region" {
  name = "my-region"
}

data "vcfa_storage_class" "sc" {
  region_id = data.vcfa_region.region.id
  name      = "My storage class"
}

resource "vcfa_content_library" "cl1" {
  org_id      = vcfa_org.test.id
  name        = "my-content-library"
  description = "Example ibrary"
  auto_attach = false # Defaults to true
  storage_class_ids = [
    data.vcfa_storage_class.sc.id
  ]
  delete_recursive = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Content Library
* `org_id` - (Optional) The reference to the Organization that the Content Library belongs to. If it is not set, assumes the
  Content Library is of type `PROVIDER`
* `delete_force` - (Optional) Defaults to `false`. On deletion, forcefully deletes the Content Library and its Content Library items. Only considered with
  `PROVIDER` Content Libraries, ignored otherwise.
* `delete_recursive` - (Optional) Defaults to `false`. On deletion, deletes the Content Library, including its Content Library items, in a single operation
* `storage_class_ids` - (Required) A set of [Storage Class IDs](/providers/vmware/vcfa/latest/docs/data-sources/storage_class) used by this Content Library
* `auto_attach` - (Optional) Defaults to `true`. For Tenant Content Libraries this field represents whether this Content Library should be
  automatically attached to all current and future namespaces in the tenant organization. If a value of `false` is supplied, then this
  Tenant Content Library will only be attached to namespaces that explicitly request it. For Provider Content Libraries this field is not needed
  for creation and will always be returned as true. This field cannot be updated after Content Library creation
* `description` - (Optional) The description of the Content Library
* `subscription_config` - (Optional) A block representing subscription settings of a Content Library:
  *  `subscription_url` - Subscription url of this Content Library
  *  `password` - Password to use to authenticate with the publisher
  *  `need_local_copy` - Whether to eagerly download content from publisher and store it locally

## Attribute Reference

* `creation_date` - The ISO-8601 timestamp representing when this Content Library was created
* `is_shared` - Whether this Content Library is shared with other Organziations
* `is_subscribed` - Whether this Content Library is subscribed from an external published library
* `library_type` - The type of content library, can be either `PROVIDER` (Content Library that is scoped to a provider) or 
  `TENANT` (Content Library that is scoped to a tenant organization)
* `version_number` - Version number of this Content library 

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. However, an experimental feature in Terraform 1.5+ allows also code generation.
See [Importing resources][importing-resources] for more information.

An existing Content Library can be [imported][docs-import] into this resource via supplying its name.
For example, using this structure, representing an existing Content Library that was **not** created using Terraform:

```hcl
resource "vcfa_content_library" "cl" {
  name = "My Already Existing Library"
}
```

You can import such Content Library into terraform state using this command

```
terraform import vcfa_content_library.cl "My Already Existing Library"
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCFA_IMPORT_SEPARATOR

After that, you can expand the configuration file and either update or delete the Content Library as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Content Library's stored properties.

[importing-resources]:https://registry.terraform.io/providers/vmware/vcfa/latest/docs/guides/importing_resources