---
page_title: "VMware Cloud Foundation Automation: Roles Management"
subcategory: ""
description: |-
 Provides guidance to roles management in VMware Cloud Foundation Automation
---

# Roles Management

-> In this document, when we mention **Tenants**, the term can be substituted with **Organizations**.

## Overview

Roles management is a combination of four entities:

- **Rights**: read-only entities, available to both provider and tenants.
- **Roles**: a container of rights that defines the privileges that can be assigned to a user. It is available to both provider and tenants.
- **Global Role**: are blueprints for roles, created in the provider, which become available as _Roles_ in the tenant.
- **Rights Bundles**: are collections of rights that define which rights become available to one or more tenants.

There are similarities among Roles, Global Roles, and Rights Bundles: all three are collections of rights for different
purposes. The similarity is in the way we create and modify these resources. We can add and remove rights to obtain a
different resource. For the purpose of describing their common functionalities, we can call these three entities **Rights Containers**.

There are also similarities between Global Roles and Rights Bundles: both resources need to be published to one or more
tenants in order to be effective. Both can become isolated if we remove all tenants, or can be maximized if we publish
to all tenants without providing a list. In this later case, the resource will be also published to future tenants.

## Rights

**Rights** ([`vcfa_right`](/providers/vmware/vcfa/latest/docs/data-sources/right)) are available as data sources. They can't be created by either provider or tenants.
They are building blocks for the other three entities (Roles, Global Roles, Rights Bundles), and can be used by simply
stating their name within the containing entity. You can also use data sources, but it would make for a crowded HCL
script, and would also increase the amount of computing needed to run a script.

Examples:

```hcl
data "vcfa_org" "org1" {
  name = "org1"
}

data "vcfa_right" "cl_item_manage" {
  name = "Content Library Item: Manage"
}

output "cl_item_manage" {
  value = data.vcfa_right.cl_item_manage
}
```

A right can have a list of **implied rights**. When such list exists, it means that, in addition to the main right, **you must
include all the implied rights** to the rights container (role, global role, rights bundle). If you don't include the
implied rights, you will get an error, listing all the rights that are missing from your entity.

## Roles

A **Role** ([`vcfa_role`](/providers/vmware/vcfa/latest/docs/resources/role)) is a set of rights that can be assigned to a user. When choosing a role for a user, we see a list of predefined
roles that are available to the organization. That list is the result of the **Global Roles** defined by the provider
and published to the tenant we are using, in addition to the roles that were created by the organization administrator.
As such, roles always belong to an organization. To define or use a role at provider level, we use the "System" organization.

## Global Roles

A **Global Role** ([`vcfa_global_role`](/providers/vmware/vcfa/latest/docs/resources/global_role)) is a definition of a role that is _published_ to one or more tenants, which in turn will see such global
roles converted into the roles they can use.
Provider can add, modify, and delete global roles. They can also alter the list of publication for each global role, to
make them available to a selected set of tenants.

## Rights Bundles

A **Rights Bundle** ([`vcfa_rights_bundle`](/providers/vmware/vcfa/latest/docs/resources/rights_bundle)) is a set of rights that can be made available to tenants. While global roles define tenant roles, a
rights bundle define which rights, independently of a global role listing, can be given to one or more tenants.

-> If you create a Rights Bundle, and it is not visible in UI (_Provider > Access Control > Rights Bundles_),
be sure to enable the `Advanced Rights Bundle Mode` feature flag in _Provider > Feature Flags_.

An example is necessary to understand the concept.

Let's say that, as a provider, you change the publishing of the rights bundle `Default Tenant Rights Bundle` and restrict its
usage to a single tenant (called `first-org`). Then, you create another rights bundle, similar to `Default Tenant Rights Bundle`,
but with only _view_ rights, and publish this bundle to another tenant (`second-org`). With this change, an Org administrator
in `first-org` will see the usual roles, with the usual sets of rights. The Org administrator in `second-role`, meanwhile,
will see the same roles, but with only half the rights, as the _managing_ rights will be missing. While this is an extreme
example, it serves to illustrate the function of rights bundles. You can create general purpose global roles for several
tenants, and then limit their reach by adding or removing rights to the rights bundle governing different tenants.

## How to include rights and implied rights into a rights container

Adding rights to one of the rights containers (Role, Global Role, Rights Bundle) is a comparable operation that works
by the same principles:

- You add an array of rights names to the `rights` field of the entity;
- If you get an error about missing implied rights, you add them to the list.

This operation is different from what we do in the UI, where when we select a right, the implied rights are added automatically.
In Terraform operations, we need to enter every right explicitly. This matter of implied rights may be confusing: rights
are not regular or implied per se: it depends on their relative status. For example, if you create a role with only
right "*Content Library Item: View*", it will work. This is the main right, and it is accepted as such.
But if you want to create a role with only right "*Content Library Item: Manage*", then it complains that you are missing
the implied right "*Content Library Item: View*". Consequently, if you know that to edit a content library you
first need to see it, you add both rights, and don't consider either of them to be implied.

For example, lets say, for the sake of simplicity, that you want to create a role with just two rights, as listed below:

```hcl
data "vcfa_org" "org1" {
  name = "org1"
}

resource "vcfa_role" "new-role" {
  org_id      = data.vcfa_org.org1.id
  name        = "new-role"
  description = "new role"
  rights = [
    "Content Library Item: Manage",
  ]
}
```

When you run `terraform apply`, you get this error:

```
╷
│ Error: The Rights set requires the following implied Rights to be added:
│ "Content Library Item: View",
│
│
│   with vcfa_role.new-role,
│   on main.tf line 448, in resource "vcfa_role" "new-role":
│  448: resource "vcfa_role" "new-role" {
│
╵
```
Thus, you update the script to include the rights mentioned in the error message

```hcl
data "vcfa_org" "org1" {
  name = "org1"
}

resource "vcfa_role" "new-role" {
  org_id      = data.vcfa_org.org1.id
  name        = "new-role"
  description = "new role"
  rights = [
    "Content Library Item: Manage",
    "Content Library Item: View",
  ]
}
```

Then repeat `terraform apply`. This time the operation succeeds.

The corresponding structure for global role and rights bundle are almost the same. You just need to add the tenants
management fields.

```hcl
resource "vcfa_global_role" "new-global-role" {
  name        = "new-global-role"
  description = "new global role"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = true
}

resource "vcfa_rights_bundle" "new-rights-bundle" {
  name        = "new-rights-bundle"
  description = "new rights bundle"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = true
}
```

## Tenant management

Rights Bundle and Global Roles have a `org_ids` section where you can list to which tenants the resource should be
published, meaning which tenants can feel the effects of this resource.

There are two fields related to managing tenants:

- `publish_to_all_orgs` with value "true" or "false".
  - If true, the resource will be published to all tenants, even if they don't exist yet. All future organizations will get to feel the benefits or restrictions published by the resource
  - If false, then we take into account the `org_ids` field.
- `org_ids` is a list of organizations (tenants) to which we want the effects of this resource to apply.

Examples:

```hcl
resource "vcfa_global_role" "new-global-role" {
  name                = "new-global-role"
  description         = "new global role"
  rights              = [/* rights list goes here */]
  publish_to_all_orgs = true
}
```
This global role will be published to all tenants, including the ones that will be created after this resource.

Now we modify it:

```hcl
data "vcfa_org" "org1" {
  name = "org1"
}

data "vcfa_org" "org2" {
  name = "org2"
}

resource "vcfa_global_role" "new-global-role" {
  name                = "new-global-role"
  description         = "new global role"
  rights              = [/* rights list goes here */]
  publish_to_all_orgs = false
  org_ids             = [data.vcfa_org.org1.id, data.vcfa_org.org2.id]
}
```

The effects of this global role are only propagated to `org1` and `org2`. Other organizations cease to see the role that
was instantiated by this global role.

Let's do another change:

```hcl
resource "vcfa_global_role" "new-global-role" {
  name                = "new-global-role"
  description         = "new global role"
  rights              = [/* rights list goes here */]
  publish_to_all_orgs = false
}
```

The `org_ids` field is removed, meaning that we don't publish to anyone. And since `publish_to_all_orgs` is false,
the tenants previously in the list are removed from publishing, making the global role isolated. It won't have
any effect on any organization until we update its tenants list.

## How to change an existing rights container

If you want to modify a Role, Global Role, or Rights Bundle that is already in your system, you need first to import
it into Terraform state, and only then you can apply your changes.

Let's say, for example, that you want to change a rights bundle `Default Tenant Rights Bundle`, to publish it only to a limited
set of tenants, while you will create a separate rights bundle for other tenants that need a different set of rights.

The import procedure works in three steps:

(1)<br>
Create a data source for the rights bundle, and a resource that takes all its attributes from the data source:

```hcl
data "vcfa_org" "org1" {
  name = "org1"
}

data "vcfa_rights_bundle" "old-rb" {
  name = "Default Tenant Rights Bundle"
}

resource "vcfa_rights_bundle" "new-rb" {
  name                = "Default Tenant Rights Bundle"
  description         = "Default set of tenant rights"
  rights              = data.vcfa_rights_bundle.old-rb.rights
  org_ids             = [data.vcfa_org.org1.id]
  publish_to_all_orgs = false
}
```

Using the data source will free you from the need of listing all the rights contained in the bundle.
It will also make the script work across different versions, where the list of rights may differ. If you were interested
in changing the rights themselves, you could add an `output` block for the data source, copy the rights to the resource
definition, and then remove or add what you need.

(2)<br>
Import the rights bundle into terraform:

```
$ terraform import vcfa_rights_bundle.new-rb "Default Tenant Rights Bundle"
```

(3)<br>
Now you can run `terraform apply`, which will remove the default condition of "publish to all tenants", replacing it
with "publish to a single tenant".

## How to clone a rights container - method 1

In the UI, there is a "clone" button that lets us create a new role, global role, or rights bundle, and then modify it.
In Terraform, we need to take a different approach, as there is no such thing as cloning a resource.
The operation requires three steps:

(1)<br>
Create a data source for the rights container, with an `output` structure that shows the full contents. For example,
to clone a global role:

```hcl
data "vcfa_global_role" "role-to-clone" {
  name = "Organization User"
}

output "role-to-clone" {
  value = data.vcfa_global_role.role-to-clone
}
```

(2)<br>
Using the data from the output, copy the rights section into a new resource

From this:

```
role-to-clone = {
  "bundle_key" = "ROLE_ORGANIZATION_USER"
  "description" = "Rights given to an organization user"
  "id" = "urn:vcloud:globalRole:b49c5a15-73fd-4390-9e87-1e1d47e69c39"
  "name" = "Organization User"
  "org_ids" = toset([
    "urn:vcloud:org:08e9ee67-0314-4157-b36c-88c686194e57",
  ])
  "publish_to_all_orgs" = true
  "read_only" = true
  "rights" = toset([
    "API Tokens: Manage",
    "Metrics: View",
    "Namespace Usage: Manage",
    "Namespace Usage: View",
    "vApp: Use Console",
  ])
}
```

to this:

```hcl
data "vcfa_org" "org1" {
  name = "org1"
}

resource "vcfa_global_role" "new-vapp-user" {
  name                = "new-cloned-role"
  description         = "New global role cloned from an existing one"
  publish_to_all_orgs = false
  rights = [
    "API Tokens: Manage",
    "Metrics: View",
    "Namespace Usage: Manage",
    "Namespace Usage: View",
    "vApp: Use Console",
  ]
  org_ids = [
    data.vcfa_org.org1.id,
  ]
}
```

(3) <br>
Remove the data source and apply the changes.

## How to clone a rights container - method 2

If you want to take one or more rights container as a basis for a new one, you may use some Terraform built-in functions
to combine sets of rights without writing down all of them.

### Example 1 - make a new global role with some rights removed from an existing one

Using [setsubtract](https://www.terraform.io/docs/language/functions/setsubtract.html) we can remove one or more items
from a given set.

```hcl
data "vcfa_global_role" "role-to-clone" {
  name = "Organization User"
}

resource "vcfa_global_role" "new-vapp-user" {
  name                = "new-cloned-role"
  description         = "New global role cloned from an existing one"
  publish_to_all_orgs = true
  rights = setsubtract(
    data.vcfa_global_role.role-to-clone.rights, # rights from existing global role
    ["vApp: Use Console"]                       # rights to be removed
  )
}
```

### Example 2 - make a new global role with a few rights more than an existing one

With the function [setunion](https://www.terraform.io/docs/language/functions/setunion.html) we can combine several
sets into one. For example, we can take the rights from both "vApp User" and "Catalog Author" into a new global role,
and if we want we can even add extra rights that we specify manually.

```hcl
data "vcfa_global_role" "role-to-clone1" {
  name = "Organization User"
}

data "vcfa_global_role" "role-to-clone2" {
  name = "Organization Auditor"
}

resource "vcfa_global_role" "super-vapp-user" {
  name                = "new-cloned-role"
  description         = "New global role cloned from an existing one"
  publish_to_all_orgs = true
  rights = setunion(
    data.vcfa_global_role.role-to-clone1.rights, # rights from existing global role
    data.vcfa_global_role.role-to-clone2.rights, # rights from existing global role
    ["Content Library: View"],                   # more rights to be added
  )
}
```
