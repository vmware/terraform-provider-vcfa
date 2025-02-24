---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_content_library_item"
sidebar_current: "docs-vcfa-resource-content-library-item"
description: |-
  Provides a VMware Cloud Foundation Automation Content Library Item resource. This can be used to manage Content Library Items.
---

# vcfa\_content\_library\_item

Provides a VMware Cloud Foundation Automation Content Library Item resource. This can be used to manage Content Library Items.

## Example Usage

```hcl
data "vcfa_content_library" "cl" {
  name = "My Library"
}

resource "vcfa_content_library_item" "cli" {
  name               = "My OVA"
  description        = "My uploaded OVA"
  content_library_id = data.vcfa_content_library.cl.id
  upload_piece_size  = 10
  file_path          = "./photon.ova"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Content Library Item
* `content_library_id` - (Required) ID of the Content Library that this Content Library Item belongs to
* `file_path` - (Required) Path to the OVA/ISO to create the Content Library Item
* `upload_piece_size` - (Optional) - When uploading the Content Library Item, this argument defines the size of the file chunks
  in which it is split on every upload request. It can possibly impact upload performance. Default 1 MB.
* `description` - (Optional) The description of the Content Library Item

## Attribute Reference

* `creation_date` - The ISO-8601 timestamp representing when this Content Library Item was created
* `item_type` - The type of Content Library Item
* `image_identifier` - Virtual Machine Identifier (VMI) of the Content Library Item. This is a ReadOnly field
* `is_published` - Whether this Content Library Item is published
* `is_subscribed` - Whether this Content Library Item is subscribed
* `last_successful_sync` - The ISO-8601 timestamp representing when this Content Library Item was last synced if subscribed
* `owner_org_id` - The reference to the organization that the Content Library Item belongs to
* `status` - Status of this Content Library Item
* `version` - The version of this Content Library Item. For a subscribed library, this version is same as in publisher library 

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. However, an experimental feature in Terraform 1.5+ allows also code generation.
See [Importing resources][importing-resources] for more information.

An existing Content Library Item can be [imported][docs-import] into this resource via supplying its name.
For example, using this structure, representing an existing Content Library Item that was **not** created using Terraform:

```hcl
resource "vcfa_content_library" "cl" {
  name = "My Already Existing Library"
}
```

You can import such Content Library Item into terraform state using the **Content Library name** and the **Item name**, with this command

```
terraform import vcfa_content_library_item.cli "My Already Existing Library"."My Already Existing Item"
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCFA_IMPORT_SEPARATOR

After that, you can expand the configuration file and either update or delete the Content Library as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Content Library Item's stored properties.

[importing-resources]:https://registry.terraform.io/providers/vmware/vcfa/latest/docs/guides/importing_resources