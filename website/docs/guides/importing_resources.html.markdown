---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: Importing resources"
sidebar_current: "docs-vcfa-guides-importing-resources"
description: |-
 Provides guidance to import resources that already exist on VMware Cloud Foundation Automation 
---

# Importing resources

Supported in Terraform *v1.5.x+*.

-> Some parts of this document describe **EXPERIMENTAL** features.

## Overview

Importing is the process of bringing a resource under Terraform control, in those cases where the resource was created
from a different agent, or a resource was created using Terraform, but some or all of its contents were not.

When we create a resource using Terraform, the side effect of this action is that the resource is stored into [Terraform state][terraform-state],
which allows us to further manipulate the resource, such as changing its contents or deleting it.

## Importing terminology

In order to import a resource, we need to issue a command, containing several elements, which are explained below.

* A **command** is a keyword of the Terraform command line tool, like the `import` in the _import command_ above.
* The resource **type** is the type of resource, such as `vcfa_org`, `vcfa_content_library`, etc.
* The **local name** (or resource **definer**) is the name given to the resource, right next to the resource type. Not to be confused with the `name`.
* The **import path** is the identification of the resource, given as the name of its ancestors, followed by the name of the resource itself.
  The import path may be different for each resource, and may include elements other than the name.
* The **id** is the same as the **import path** in Terraform parlance. When we see the `id` mentioned in Terraform import
  documentation, it is talking about the import path.
* The **resource block** is the portion of HCL code that defines the resource.
* The **name** is the unequivocal name of the resource, independently of the HCL script.
* The **separator** is a character used to separate the elements of the `import path`. The default is the dot (`.`).
* The **Terraform state** is the representation about our resources as understood by Terraform.
* A **stage** is a part of the Terraform workflow, which is mainly one of `plan`, `apply`, `destroy`, `import`. Each one of these
  has a corresponding `command`, but not all Terraform commands have a stage.

For example:

```hcl
resource "vcfa_org_local_user" "admin_org_user" {
  org  = "my-org"
  name = "philip"
}
```

```shell
terraform import vcfa_org_local_user.admin_org_user my-org.philip
```

In the two snippets above:

* `vcfa_org_local_user` is the resource **type**
* `admin_org_user` is the resource **definer**
* `import` is the **command**
* `.` is the **separator**
* `philip` is the resource **name**
* `my-org.philip` is the **resource path** or **id**
* All 5 lines of HCL code starting from `resource` are the **resource block**
* The **Terraform state** (not visible in the above script) is collected in the file `terraform.tfstate`

## Basic importing

Up to Terraform 1.4.x, importing meant the conjunction of two operations:
1. Writing the resource definition into an HCL script
2. Running the command below, also known as "**the import command**"

```
terraform import vcfa_resource_type.resource_definer path_to_resource
```

The effect of the above actions (which we can also perform in later versions of Terraform) is that the resource is
imported into the [state][terraform-state].
The drawback of this approach is that we need to write the HCL definition of the resource manually, which could result
in a very time-consuming operation.

## Import mechanics

When we run a `terraform import` command like the one in the previous section, Terraform will try to read all the
of the resource and fill the `state` with the resource information.
That completes the **import** stage, but it doesn't mean that the code is usable from now on.
In fact, running `terraform plan` after the import, would result in an error.

```
╷
│ Error: Missing required argument
│
│   on config.tf line 40, in resource "vcfa_org_local_user" "admin_org_user":
│   40: resource "vcfa_org_local_user" "admin_org_user" {
│
│ The argument "role_ids" is required, but no definition was found.
```

Which means that we need to edit the HCL script, and add all the necessary elements that are missing. We may use the
data from the state file (`terraform.tfstate`) to supply the missing properties.

## Semi-Automated import (Terraform v1.5+)

~> Terraform warns that this procedure is considered **experimental**.

Terraform v1.5 introduces the concept of an [import block][terraform-import], which replaces the `import` command.
Instead of 

```shell
terraform import vcfa_org_local_user.admin_org_user my-org.philip
```

we would put the import instructions directly into the HCL script

```hcl
import {
  to = vcfa_org_local_user.admin_org_user
  id = "my-org.philip"
}
```

There are two differences between the old and new import methods:
* the import happens as part of the `apply` stage, rather than on a separate command;
* Although we could write the resource block ourselves, we can now generate the HCL code using a Terraform command.

To generate the source HCL, we issue the command

```shell
terraform plan -generate-config-out=generated_resources.tf
```

The above command, and the next `terraform plan` or `terraform apply` will show one more set of actions to perform

```
 10 to import, 0 to add, 9 to change, 0 to destroy.
```

Here we see that the import is an operation that will happen during `apply`.
  
## Troubleshooting

-> Since we refer to an experimental feature, issues and relative advice given in this section may change in future
  releases or get fixed due to upstream improvements.

### Required field not found

Some resources require several properties to be filled. For example, when creating a VDC group, we need to indicate
which of the participating VDCs is the starting one.

Let's try to import an existing VDC group:

```hcl
import {
  to = vcfa_vdc_group.vdc-group-datacloud
  id = "datacloud.vdc-group-datacloud"
}
```
In this file we are saying that we want to import the VDC group `vdc-group-datacloud`, belonging to the organization `datacloud`.

```shell
terraform plan -generate-config-out=generated_resources.tf
```
```
data.vcfa_resource_list.vdc-groups: Reading...
vcfa_vdc_group.vdc-group-datacloud: Preparing import... [id=datacloud.vdc-group-datacloud]
data.vcfa_resource_list.vdc-groups: Read complete after 2s [id=list-vdc-groups]
vcfa_vdc_group.vdc-group-datacloud: Refreshing state... [id=urn:vcloud:vdcGroup:db815539-c885-4d9b-9992-aac82dce89d0]

Planning failed. Terraform encountered an error while generating this plan.
[...]
╷
│ Error: Missing required argument
│
│   with vcfa_vdc_group.vdc-group-datacloud,
│   on generated_resources.tf line 8:
│   (source code not available)
│
│ The argument "starting_vdc_id" is required, but no definition was found.
╵
```

The Terraform interpreter signals that there is one missing property. Since the current syntax of import blocks does
not allow any adjustments, the only possible workaround is to update the generated HCL code. Fortunately, the above error
does not prevent the generation of the code.
If we edit the file `generated_resources.tf`, changing the value for `starting_vdc_id` from
`null` to the ID of the first VDC, the import will succeed.

### Phantom updates

In addition to missing required properties, we may have the problem of properties that are needed during creation, but
their values are not stored in the vcfa, and consequently can't be retrieved and used to populate the importing HCL code.
For example, the [`accept_all_eulas`][accept-all-eulas] property is only used during VM creation, but we can't retrieve it
from the VM data in vcfa.
When we have such fields, Terraform will signal that the resource needs to be updated, and it will do so at the next
occurrence of `terraform apply`. This is a minor annoyance, which will delay the operation by a few seconds, but which
won't actually change anything in the resource. What this update means is that Terraform is trying to match the HCL data
with the resource stored data. We won't be making any real changes in the VM: nonetheless, we should be vigilant and
make sure that the updates being proposed don't touch important data. If they do, we should probably edit the generated
code and set the record straight.

### Lack of dependency ties

The code generation is good enough to put the resource information into Terraform state, but it won't write the HCL code
the same way we would. Most importantly, the names or IDs of other resources will be placed verbatim into the resource
definition, rather than using a reference.

For example, when creating a vApp template, we may write the following:

```hcl
resource "vcfa_catalog_vapp_template" "my_template" {
  catalog_id  = vcfa_catalog.mycatalog.id
  name        = "my_template"
  description = "my template"
}
```
However, the corresponding generated code would be:

```hcl
resource "vcfa_catalog_vapp_template" "my_template" {
  catalog_id  = "urn:vcloud:catalog:59b15c74-8dea-4331-ae2c-4fc4217c4191"
  name        = "my_template"
  description = "my template"
}
```

This code would work well when we use it to update the vApp template, but it may become a problem when we want to delete
all the resources. The lack of dependency information will cause the removal to happen in a random order, and we may
see "entity not found" errors during such operations. For example, if the catalog deletion happens before the vApp template
deletion, the template will not exist by the time Terraform attempts to retrieve it for removal.

There is no simple solution to this issue, other than manually editing the generated HCL code to add dependency instructions.

## Examples

There are two complete examples of multiple resource imports in the [`terraform-provider-vcfa` repository][examples].
They show how we can import multiple VMs, or multiple catalog items, with step-by-step instructions.

[terraform-state]:https://developer.hashicorp.com/terraform/language/state
[terraform-import]:https://developer.hashicorp.com/terraform/language/import
[vcfa-resource-list]:https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/resource_list
[vcfa-cloned-vapp]:https://registry.terraform.io/providers/vmware/vcfa/3.10.0/docs/resources/cloned_vapp
[accept-all-eulas]:https://registry.terraform.io/providers/vmware/vcfa/3.10.0/docs/resources/vapp_vm#accept_all_eulas
[examples]:https://github.com/vmware/terraform-provider-vcfa/tree/import-compound-resources/examples/importing