---
page_title: "VMware Cloud Foundation Automation: vcfa_org_regional_networking_vpc_qos"
description: |-
  Provides a data source to read VMware Cloud Foundation Automation Organization Regional Networking VPC QoS settings.
---

# vcfa_org_regional_networking_vpc_qos

Provides a data source to read [Organization Regional Networking][vcfa_org_regional_networking-ds] VPC QoS settings in VMware Cloud Foundation Automation.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "my-org-name"
}

data "vcfa_org_regional_networking" "demo" {
  name   = "my-name"
  org_id = vcfa_org.demo.id
}

data "vcfa_org_regional_networking_vpc_qos" "demo" {
  org_regional_networking_id = vcfa_org_regional_networking.demo.id
}
```

## Argument Reference

The following arguments are supported:

- `org_regional_networking_id` - (Required) The ID of [Organization Regional Networking][vcfa_org_regional_networking-ds] configuration

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_regional_networking_vpc_qos`](/providers/vmware/vcfa/latest/docs/resources/org_regional_networking_vpc_qos)
resource are available.

[vcfa_org_regional_networking-ds]: /providers/vmware/vcfa/latest/docs/data-sources/org_regional_networking