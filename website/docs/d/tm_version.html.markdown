---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_tm_version"
sidebar_current: "docs-data-source-vcfa-tm-version"
description: |-
  Provides a data source to fetch the Tenant Manager version from VMware Cloud Foundation Automation,
  its maximum supported API version and perform some optional checks with version constraints.
---

# vcfa\_tm\_version

Provides a data source to fetch the Tenant Manager version from VMware Cloud Foundation Automation, its maximum supported API version and
perform some optional checks with version constraints.

Requires System Administrator privileges.

## Example Usage

```hcl
# This data source will assert that the VCFA Tenant Manager version is exactly 10.5.1, otherwise it will fail
data "vcfa_tm_version" "eq_1051" {
  condition         = "= 10.5.1"
  fail_if_not_match = true
}

# This data source will assert that the VCFA Tenant Manager version is greater than or equal to 10.4.2, but it won't fail if it is not
data "vcfa_tm_version" "gte_1042" {
  condition         = ">= 10.4.2"
  fail_if_not_match = false
}

output "is_gte_1042" {
  value = data.vcfa_tm_version.gte_1042.matches_condition # Will show false if we're using a VCFA Tenant Manager version < 10.4.2
}

# This data source will assert that the VCFA Tenant Manager version is less than 10.5.0
data "vcfa_tm_version" "lt_1050" {
  condition         = "< 10.5.0"
  fail_if_not_match = true
}

# This data source will assert that the VCFA Tenant Manager version is 10.5.X
data "vcfa_tm_version" "is_105" {
  condition         = "~> 10.5"
  fail_if_not_match = true
}

# This data source will assert that the VCFA Tenant Manager version is not 10.5.1
data "vcfa_tm_version" "not_1051" {
  condition         = "!= 10.5.1"
  fail_if_not_match = true
}

# Output the version and API version of Tenant Manager
data "vcfa_tm_version" "version" {}

output "tenant_manager_version" {
  value = data.vcfa_tm_version.version.tm_version
}

output "tenant_manager_api_version" {
  value = data.vcfa_tm_version.version.tm_api_version
}
```

## Argument Reference

The following arguments are supported:

- `condition` - (Optional) A version constraint to check against the VCFA Tenant Manager version
- `fail_if_not_match` - (Optional) Required if `condition` is set. Throws an error if the version constraint set in `condition` is not met.
  Defaults to `false`

## Attribute Reference

- `matches_condition` - It is true if the VCFA Tenant Manager version matches the constraint set in `condition`
- `tm_version` - The VCFA Tenant Manager version
- `tm_api_version` - The maximum supported VCFA Tenant Manager API version
