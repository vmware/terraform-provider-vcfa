---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_region"
sidebar_current: "docs-vcfa-data-source-region"
description: |-
  Provides a data source to read Regions in VMware Cloud Foundation Automation.
---

# vcfa\_region

Provides a data source to read Regions in VMware Cloud Foundation Automation.

## Example Usage

```hcl
data "vcfa_region" "one" {
  name = "region-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name of existing Region

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_region`](/providers/vmware/vcfa/latest/docs/resources/region) resource are available.
