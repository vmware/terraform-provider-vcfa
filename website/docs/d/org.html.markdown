---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org"
sidebar_current: "docs-vcfa-data-source-org"
description: |-
  Provides a data source to read VMware Cloud Foundation Automation Organizations.
---

# vcfa\_org

Provides a data source to read VMware Cloud Foundation Automation Organizations.

## Example Usage

```hcl
data "vcfa_org" "existing" {
  name = "my-org-name"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of organization.

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org`](/providers/vmware/vcfa/latest/docs/resources/org) resource are available.