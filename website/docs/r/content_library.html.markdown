---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_content_library"
sidebar_current: "docs-vcfa-resource-content-library"
description: |-
  Provides a resource to manage Content Libraries in VMware Cloud Foundation Automation. It can be used to upload
  Content Library Items such as ISO files, OVAs and OVFs.
---

# vcfa\_content\_library

Provides a resource to manage Content Libraries in VMware Cloud Foundation Automation. It can be used to upload
[Content Library Items][vcfa_content_library_item] such as ISO files, OVAs and OVFs.

## Example Usage for a Provider Content Library

The snippet below will create a Content Library of type `PROVIDER`. To achieve that, one needs to
read the `System` (Provider) organization with the [`vcfa_org` data source][vcfa_org-ds] to
use it as Organization reference in the `vcfa_content_library` resource.

Note that the snippet assumes that the [Region][vcfa_region-ds] is already created.

```hcl
data "vcfa_org" "system" {
  name = "System"
}

data "vcfa_region" "region" {
  name = "My Region"
}

data "vcfa_storage_class" "sc" {
  region_id = data.vcfa_region.region.id
  name      = "vSAN Default Storage Policy"
}

resource "vcfa_content_library" "cl" {
  org_id      = data.vcfa_org.system.id
  name        = "My Library"
  description = "A simple library"
  storage_class_ids = [
    data.vcfa_storage_class.sc.id
  ]
}

resource "vcfa_content_library" "cl2" {
  org_id = data.vcfa_org.system.id
  name   = "My Subscribed Library"
  # Subscribed libraries inherit description from publisher
  storage_class_ids = [
    data.vcfa_storage_class.sc.id
  ]
  subscription_config {
    subscription_url = "https://my-vcenter.com/cls/vcsp/lib/41eb97db-e1b4-47e6-b0f3-5e02aa3830f7/lib.json"
    need_local_copy  = true
  }
}
```

## Example Usage for a Tenant Content Library as a System Administrator

The snippet below will create a Content Library of type `TENANT` but logged in as System Administrator. To achieve that, one needs to
configure the System Administrator credentials in the provider configuration block, and read the target organization with a
[`vcfa_org`][vcfa_org-ds] data source to use it as Organization reference in the `vcfa_content_library` resource.

Note that the snippet assumes that the [vCenter][vcfa_vcenter-ds], [NSX][vcfa_nsx_manager-ds], [Region][vcfa_region-ds], etc.
are already created in the System (Provider) organization.

The snippet also creates a [Region Quota][vcfa_region_quota], required to host Content Libraries in the
Organizations.

~> To create subscribed libraries (with `subscription_config` block), check that the [`vcfa_org_settings`](/providers/vmware/vcfa/latest/docs/resources/org_settings)
of the target Organization allows it.

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
  depends_on = [vcfa_org_region_quota.test]
}
```

## Example Usage for a Tenant Content Library as a Tenant User

The snippet below will create a Content Library of type `TENANT` logged in as a regular tenant user. To achieve that, one needs to
configure the tenant user credentials in the provider configuration block, and read the target organization with a
[`vcfa_org`][vcfa_org-ds] data source to use it as Organization reference in the `vcfa_content_library` resource.

~> To create subscribed libraries (with `subscription_config` block), check that the [`vcfa_org_settings`][vcfa_org_settings]
of the target Organization allows it.

```hcl
data "vcfa_org" "org" {
  name = "my-org"
}

data "vcfa_region" "region" {
  name = "my-region"
}

data "vcfa_storage_class" "sc" {
  region_id = data.vcfa_region.region.id
  name      = "My storage class"
}

resource "vcfa_content_library" "cl1" {
  org_id      = vcfa_org.org.id
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
* `org_id` - (Required) The reference to the Organization that the Content Library belongs to.  For Content Libraries of type `PROVIDER`,
  a reference to the `System` org with [`vcfa_org`][vcfa_org-ds] data source must be provided
* `delete_force` - (Optional) Defaults to `false`. On deletion, forcefully deletes the Content Library and its Content Library items. Only considered with
  `PROVIDER` Content Libraries, ignored otherwise
* `delete_recursive` - (Optional) Defaults to `false`. On deletion, deletes the Content Library, including its Content Library items, in a single operation
* `storage_class_ids` - (Required) A set of [Storage Class IDs][vcfa_storage_class-ds] used by this Content Library
* `auto_attach` - (Optional) Defaults to `true`. For `TENANT` Content Libraries this field represents whether this Content Library should be
  automatically attached to all current and future namespaces in the Organization. If a value of `false` is supplied, then this
  Tenant Content Library will only be attached to namespaces that explicitly request it. For `PROVIDER` Content Libraries this field is not needed
  for creation and will always be returned as `true`. This field cannot be updated after creation
* `description` - (Optional) The description of the Content Library. Not used if the library is subscribed to another one (see `subscription_config` below), as
  the value will be the one from publisher library
* `subscription_config` - (Optional) A block representing subscription settings of a Content Library:
  *  `subscription_url` - Subscription url of this Content Library
  *  `password` - Password to use to authenticate with the publisher
  *  `need_local_copy` - Whether to eagerly download content from publisher and store it locally

~> To use `subscription_config` block in `TENANT` type Content Libraries, check that the [`vcfa_org_settings`][vcfa_org_settings]
of the target Organization allows it.

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
For example, using this structure, representing an existing `TENANT` Content Library that was **not** created using Terraform:

```hcl
data "vcfa_org" "org" {
  name = "my-org"
}

resource "vcfa_content_library" "cl" {
  org_id = data.vcfa_org.org.id
  name   = "My Already Existing Library"
}
```

You can import such Content Library into terraform state using this command:

```
terraform import vcfa_content_library.cl "my-org"."My Already Existing Library"
```

For an existing `PROVIDER` Content Library that was **not** created using Terraform:


```hcl
data "vcfa_org" "org" {
  name = "System"
}

resource "vcfa_content_library" "cl" {
  org_id = data.vcfa_org.org.id
  name   = "My Already Existing Library"
}
```

You can import such Content Library into terraform state using this command:

```
terraform import vcfa_content_library.cl "System"."My Already Existing Library"
```

NOTE: the default separator (.) can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

After that, you can expand the configuration file and either update or delete the Content Library as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Content Library's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_content_library_item]: /providers/vmware/vcfa/latest/docs/resources/content_library_item
[vcfa_nsx_manager-ds]: /providers/vmware/vcfa/latest/docs/data-sources/nsx_manager
[vcfa_org-ds]: /providers/vmware/vcfa/latest/docs/data-sources/org
[vcfa_org_settings]: /providers/vmware/vcfa/latest/docs/resources/org_settings
[vcfa_region-ds]: /providers/vmware/vcfa/latest/docs/data-sources/region
[vcfa_region_quota]: /providers/vmware/vcfa/latest/docs/resources/region_quota
[vcfa_storage_class-ds]: /providers/vmware/vcfa/latest/docs/data-sources/storage_class
[vcfa_vcenter-ds]: /providers/vmware/vcfa/latest/docs/data-sources/vcenter