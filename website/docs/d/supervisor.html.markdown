---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor"
sidebar_current: "docs-vcfa-data-source-supervisor"
description: |-
  Provides a data source to read a Supervisor in VMware Cloud Foundation Automation.
---

# vcfa\_supervisor

Provides a data source to read a Supervisor in VMware Cloud Foundation Automation.

~> This data source can only be used by **System Administrators**

## Example Usage

```hcl
data "vcfa_vcenter" "one" {
  name = "vcenter-one"
}

data "vcfa_supervisor" "one" {
  name       = "my-supervisor-name"
  vcenter_id = data.vcfa_vcenter.one.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Supervisor
* `vcenter_id` - (Required) vCenter server ID that contains this Supervisor

## Attribute Reference

* `region_id` - Region ID that consumes this Supervisor
