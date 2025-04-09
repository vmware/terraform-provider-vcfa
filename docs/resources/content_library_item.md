---
page_title: "VMware Cloud Foundation Automation: vcfa_content_library_item"
subcategory: ""
description: |-
  Provides a resource to manage Content Library Items in VMware Cloud Foundation Automation. Allows to upload an ISO file, an OVA or an OVF
  to a Content Library.
---

# vcfa_content_library_item

Provides a resource to manage Content Library Items in VMware Cloud Foundation Automation. Allows to upload an ISO file, an OVA or an OVF
to a [Content Library][vcfa_content_library].

_Used by: **Provider**, **Tenant**_

## Example Usage

The target [Content Library][vcfa_content_library] can be located inside an [Organization][vcfa_org] or in the System (Provider) org.

When uploading an OVF file, be sure that all required inner elements are specified:

```hcl
data "vcfa_content_library" "cl" {
  name = "My Library"
}

resource "vcfa_content_library_item" "ova" {
  name               = "my-ova"
  description        = "Description of my-ova"
  content_library_id = vcfa_content_library.cl.id
  file_paths         = ["./my_ova.ova"]
}

resource "vcfa_content_library_item" "iso" {
  name               = "iso1"
  description        = "Description of iso1"
  content_library_id = vcfa_content_library.cl.id
  file_paths         = ["./linux.iso"]
}

resource "vcfa_content_library_item" "ovf" {
  name               = "ovf"
  description        = "Description of OVF"
  content_library_id = vcfa_content_library.cl.id
  file_paths         = ["./my-ovf/descriptor.ovf", "./my-ovf/disk1.vmdk"]
}

```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the Content Library Item
- `content_library_id` - (Required) ID of the [Content Library][vcfa_content_library] that this Content Library Item belongs to
- `file_paths` - (Required) A single path to an OVA/ISO, or multiple paths for an OVF and its referenced files, to create the Content Library Item
- `upload_piece_size` - (Optional) - When uploading the Content Library Item, this argument defines the size of the file chunks
  in which it is split on every upload request. It can possibly impact upload performance. Default 1 MB
- `description` - (Optional) The description of the Content Library Item

## Attribute Reference

- `creation_date` - The ISO-8601 timestamp representing when this Content Library Item was created
- `item_type` - The type of Content Library Item
- `image_identifier` - Virtual Machine Identifier (VMI) of the Content Library Item. This is a read-only field
- `is_published` - Whether this Content Library Item is published
- `is_subscribed` - Whether this Content Library Item is subscribed
- `last_successful_sync` - The ISO-8601 timestamp representing when this Content Library Item was last synced if subscribed
- `owner_org_id` - The reference to the organization that the Content Library Item belongs to
- `status` - Status of this Content Library Item
- `version` - The version of this Content Library Item. For a subscribed library, this version is same as in publisher library

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. However, an experimental feature in Terraform 1.5+ allows also code generation.
See [Importing resources][importing-resources] for more information.

An existing Content Library Item can be [imported][docs-import] into this resource via supplying its Organization name, Content Library name
and Item name. For example, using this structure, representing an existing Content Library Item that was **not** created using Terraform:

```hcl
resource "vcfa_content_library" "cl" {
  name = "My Already Existing Library"
}
```

You can import such Content Library Item into terraform state:

```
terraform import vcfa_content_library_item.cli "My existing Org"."My Already Existing Library"."My Already Existing Item"
```

If the Content Library Item is a `PROVIDER` one (System org):

```
terraform import vcfa_content_library_item.cli System."My Already Existing Library"."My Already Existing Item"
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

After that, you can expand the configuration file and either update or delete the Content Library Item as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Content Library Item's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_content_library]: /providers/vmware/vcfa/latest/docs/resources/content_library
[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org
