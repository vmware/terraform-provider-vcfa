# Greenfield configuration for VMware Cloud Foundation Automation

In this folder you can find configuration files (`.tf`) to apply to a completely fresh installation of VMware Cloud Foundation Automation.

It is divided into two parts, corresponding to two main roles:

- [`provider`](provider): Is the part that the VMware Cloud Foundation Automation administrator will apply to create the
  required Organizations (tenants) and configure the underlying infrastructure (vCenter, NSX, Provider Gateways, roles...)
- [`tenant`](tenant): Is the part that a tenant user will apply to configure the layout of the Organization, like Content Libraries