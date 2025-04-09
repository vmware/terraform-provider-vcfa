---
page_title: "VMware Cloud Foundation Automation: Importing resources"
subcategory: ""
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

- A **command** is a keyword of the Terraform command line tool, like the `import` in the _import command_ above.
- The resource **type** is the type of resource, such as `vcfa_org`, `vcfa_content_library`, etc.
- The **local name** (or resource **definer**) is the name given to the resource, right next to the resource type. Not to be confused with the `name`.
- The **import path** is the identification of the resource, given as the name of its ancestors, followed by the name of the resource itself.
  The import path may be different for each resource, and may include elements other than the name.
- The **id** is the same as the **import path** in Terraform parlance. When we see the `id` mentioned in Terraform import
  documentation, it is talking about the import path.
- The **resource block** is the portion of HCL code that defines the resource.
- The **name** is the unequivocal name of the resource, independently of the HCL script.
- The **separator** is a character used to separate the elements of the `import path`. The default is the dot (`.`).
- The **Terraform state** is the representation about our resources as understood by Terraform.
- A **stage** is a part of the Terraform workflow, which is mainly one of `plan`, `apply`, `destroy`, `import`. Each one of these
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

- `vcfa_org_local_user` is the resource **type**
- `admin_org_user` is the resource **definer**
- `import` is the **command**
- `.` is the **separator**
- `philip` is the resource **name**
- `my-org.philip` is the **resource path** or **id**
- All 5 lines of HCL code starting from `resource` are the **resource block**
- The **Terraform state** (not visible in the above script) is collected in the file `terraform.tfstate`

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
- the import happens as part of the `apply` stage, rather than on a separate command;
- Although we could write the resource block ourselves, we can now generate the HCL code using a Terraform command.

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

[terraform-state]:https://developer.hashicorp.com/terraform/language/state
[terraform-import]:https://developer.hashicorp.com/terraform/language/import