## 1.1.1 (Unreleased)

Changes in progress for v1.1.1 are available at [.changes/v1.1.1](https://github.com/vmware/terraform-provider-vcfa/tree/main/.changes/v1.1.1) until the release.

## 1.1.0 (May 19, 2026)

### FEATURES
- Add support for VCFA 9.1 ([#129](https://github.com/vmware/terraform-provider-vcfa/pull/129))
- Add support for vcfa_supervisor_namespace resource updates ([#156](https://github.com/vmware/terraform-provider-vcfa/pull/156))
- **New Data Source:** `vcfa_shared_subnet` to read Shared Subnets ([#168](https://github.com/vmware/terraform-provider-vcfa/pull/168))
- **New Resource:** `vcfa_shared_subnet` to manage Shared Subnets ([#168](https://github.com/vmware/terraform-provider-vcfa/pull/168))
- **New Data Source:** `vcfa_distributed_vlan_connection` to read Distributed VLAN Connections ([#170](https://github.com/vmware/terraform-provider-vcfa/pull/170))
- **New Resource:** `vcfa_distributed_vlan_connection` to manage Distributed VLAN Connections ([#170](https://github.com/vmware/terraform-provider-vcfa/pull/170))

### IMPROVEMENTS
- Update `vcfa_supervisor_namespace` resource and datasource to support the latest `v1alpha3` CCI changes [[#153](https://github.com/vmware/terraform-provider-vcfa/pull/153)]. Added the new attributes `conditions`, `content_libraries`, `content_sources_class_config_overrides`, `infra_policies`, `infra_policy_names`, `seg_name`, `shared_subnet_names`, `storage_classes`, `vm_classes`, `vm_classes_class_config_overrides`, and `zones`.
- Update `vcfa_ip_space` datasource and resource [[#165](https://github.com/vmware/terraform-provider-vcfa/pull/165)]. Added the new attributes `cidr_blocks`, `ip_address_ranges`, `provider_visibility_only`, `reserved_ip_address_ranges`, `backing_id`, `is_imported_ip_block`, and `subnet_exclusive`.
- Update `vcfa_provider_gateway` datasource and resource [[#165](https://github.com/vmware/terraform-provider-vcfa/pull/165)]. Added the new attributes `inbound_remote_networks`, `allow_advertising_private_ip_blocks`, `nat_config_enabled`, `nat_config_ip_space_id`, `nat_config_logging`, and `gateway_connection_backing_id`.
- Add an example at the `vcfa_org_oidc` resource on how to auto-generate signature keys for the OIDC flow ([#167](https://github.com/vmware/terraform-provider-vcfa/pull/167))
- Update `vcfa_org_settings` datasource and resource properties to add the new optional attribute `can_subscribe_to_third_party_libraries` ([#173](https://github.com/vmware/terraform-provider-vcfa/pull/173))
- Update `vcfa_content_library` datasource and resource to allow reading and managing Project Scoped Content Libraries [[#180](https://github.com/vmware/terraform-provider-vcfa/pull/180)]. Added the new attributes `is_project_scoped`. `all_projects_permission`, and `project_permissions`.

### BUG FIXES
- Fix filtering by `region_id` at the `vcfa_region_storage_policy` and `vcfa_storage_class` datasources ([#166](https://github.com/vmware/terraform-provider-vcfa/pull/166))

### DEPRECATIONS
- Deprecate `storage_classes_initial_class_config_overrides` in favor of `storage_classes_class_config_overrides` and `zones_initial_class_config_overrides` in favor of `zones_class_config_overrides` in `vcfa_supervisor_namespace` resource and datasource ([#153](https://github.com/vmware/terraform-provider-vcfa/pull/153))
- Deprecate `internal_scope` in favor of `cidr_blocks` and `external_scope` in favor of `inbound_remote_networks` in `vcfa_ip_space` resource and datasource [[#165](https://github.com/vmware/terraform-provider-vcfa/pull/165)].

### NOTES
- Increase sleep for `TestAccVcfaSupervisorNamespace` test so it does not fail intermittently ([#104](https://github.com/vmware/terraform-provider-vcfa/pull/104), [#105](https://github.com/vmware/terraform-provider-vcfa/pull/105))
- Amend IP Space names in tests to be compatible with Kubernetes naming convention ([#118](https://github.com/vmware/terraform-provider-vcfa/pull/118), [#119](https://github.com/vmware/terraform-provider-vcfa/pull/119))
- Fix Supervisor Namespace tests by setting the proper dependency on region network settings ([#144](https://github.com/vmware/terraform-provider-vcfa/pull/144))
- Make VPC name configurable in SupervisorNamespace tests ([#146](https://github.com/vmware/terraform-provider-vcfa/pull/146))
- Update tests to dynamically get the Edge Cluster name ([#157](https://github.com/vmware/terraform-provider-vcfa/pull/157))
- Bump golang and dependencies ([#192](https://github.com/vmware/terraform-provider-vcfa/pull/192))

## 1.0.0 (June 18, 2025)

### FEATURES

- **New Data Source:** `vcfa_version` to read version details from VCFA ([#1](https://github.com/vmware/terraform-provider-vcfa/pull/1), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66))
- **New Resource:** `vcfa_org` to manage Organizations ([#3](https://github.com/vmware/terraform-provider-vcfa/pull/3))
- **New Data Source:** `vcfa_org` to read Organizations ([#3](https://github.com/vmware/terraform-provider-vcfa/pull/3))
- **New Resource:** `vcfa_nsx_manager` to manage VMware Cloud Foundation Automation NSX Managers
  ([#4](https://github.com/vmware/terraform-provider-vcfa/pull/4), [#33](https://github.com/vmware/terraform-provider-vcfa/pull/33), [#55](https://github.com/vmware/terraform-provider-vcfa/pull/55))
- **New Data Source:** `vcfa_nsx_manager` to read VMware Cloud Foundation Automation NSX Managers
  ([#4](https://github.com/vmware/terraform-provider-vcfa/pull/4), [#33](https://github.com/vmware/terraform-provider-vcfa/pull/33))
- **New Resource:** `vcfa_region` to manage Regions ([#5](https://github.com/vmware/terraform-provider-vcfa/pull/5), [#11](https://github.com/vmware/terraform-provider-vcfa/pull/11), [#34](https://github.com/vmware/terraform-provider-vcfa/pull/34), [#51](https://github.com/vmware/terraform-provider-vcfa/pull/51))
- **New Data Source:** `vcfa_region` to read Regions ([#5](https://github.com/vmware/terraform-provider-vcfa/pull/5))
- **New Data Source:** `vcfa_supervisor` to read Supervisors ([#5](https://github.com/vmware/terraform-provider-vcfa/pull/5))
- **New Data Source:** `vcfa_supervisor_zone` to read Supervisor Zones ([#5](https://github.com/vmware/terraform-provider-vcfa/pull/5))
- **New Resource:** `vcfa_ip_space` to manage IP Spaces ([#8](https://github.com/vmware/terraform-provider-vcfa/pull/8), [#39](https://github.com/vmware/terraform-provider-vcfa/pull/39), [#40](https://github.com/vmware/terraform-provider-vcfa/pull/40), [#51](https://github.com/vmware/terraform-provider-vcfa/pull/51))
- **New Data Source:** `vcfa_ip_space` to read IP Spaces ([#8](https://github.com/vmware/terraform-provider-vcfa/pull/8))
- **New Resource:** `vcfa_org_region_quota` to manage Organization Region Quotas ([#9](https://github.com/vmware/terraform-provider-vcfa/pull/9), [#30](https://github.com/vmware/terraform-provider-vcfa/pull/30), [#31](https://github.com/vmware/terraform-provider-vcfa/pull/31))
- **New Data Source:** `vcfa_org_region_quota` to read Organization Region Quotas ([#9](https://github.com/vmware/terraform-provider-vcfa/pull/9), [#31](https://github.com/vmware/terraform-provider-vcfa/pull/31))
- **New Data Source:** `vcfa_region_zone` to read Region Zones ([#9](https://github.com/vmware/terraform-provider-vcfa/pull/9))
- **New Resource:** `vcfa_provider_gateway` to manage Provider Gateways ([#10](https://github.com/vmware/terraform-provider-vcfa/pull/10), [#69](https://github.com/vmware/terraform-provider-vcfa/pull/69), [#99](https://github.com/vmware/terraform-provider-vcfa/pull/99))
- **New Data Source:** `vcfa_provider_gateway` to read Provider Gateways ([#10](https://github.com/vmware/terraform-provider-vcfa/pull/10), [#69](https://github.com/vmware/terraform-provider-vcfa/pull/69))
- **New Data Source:** `vcfa_tier0_gateway` to read available Tier-0 Gateways from NSX ([#10](https://github.com/vmware/terraform-provider-vcfa/pull/10))
- **New Data Source:** `vcfa_edge_cluster` to read and sync Edge Clusters ([#11](https://github.com/vmware/terraform-provider-vcfa/pull/11), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66))
- **New Resource:** `vcfa_edge_cluster_qos` to manage QoS settings for Edge Clusters ([#11](https://github.com/vmware/terraform-provider-vcfa/pull/11), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66))
- **New Data Source:** `vcfa_edge_cluster_qos` to read QoS settings for Edge Clusters ([#11](https://github.com/vmware/terraform-provider-vcfa/pull/11), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66))
- **New Resource:** `vcfa_content_library` to manage Content Libraries ([#12](https://github.com/vmware/terraform-provider-vcfa/pull/12), [#42](https://github.com/vmware/terraform-provider-vcfa/pull/42), [#43](https://github.com/vmware/terraform-provider-vcfa/pull/43), [#50](https://github.com/vmware/terraform-provider-vcfa/pull/50), [#53](https://github.com/vmware/terraform-provider-vcfa/pull/53), [#62](https://github.com/vmware/terraform-provider-vcfa/pull/62), [#69](https://github.com/vmware/terraform-provider-vcfa/pull/69), [#72](https://github.com/vmware/terraform-provider-vcfa/pull/72))
- **New Data Source:** `vcfa_content_library` to read Content Libraries ([#12](https://github.com/vmware/terraform-provider-vcfa/pull/12), [#42](https://github.com/vmware/terraform-provider-vcfa/pull/42), [#50](https://github.com/vmware/terraform-provider-vcfa/pull/50), [#53](https://github.com/vmware/terraform-provider-vcfa/pull/53), [#62](https://github.com/vmware/terraform-provider-vcfa/pull/62), [#69](https://github.com/vmware/terraform-provider-vcfa/pull/69), [#70](https://github.com/vmware/terraform-provider-vcfa/pull/70), [#72](https://github.com/vmware/terraform-provider-vcfa/pull/72))
- **New Data Source:** `vcfa_region_storage_policy` to read Region Storage Policies ([#12](https://github.com/vmware/terraform-provider-vcfa/pull/12))
- **New Data Source:** `vcfa_storage_class` to read Storage Classes ([#12](https://github.com/vmware/terraform-provider-vcfa/pull/12))
- **New Resource:** `vcfa_content_library_item` to manage Content Library Items ([#13](https://github.com/vmware/terraform-provider-vcfa/pull/13), [#46](https://github.com/vmware/terraform-provider-vcfa/pull/46), [#50](https://github.com/vmware/terraform-provider-vcfa/pull/50), [#53](https://github.com/vmware/terraform-provider-vcfa/pull/53), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66), [#75](https://github.com/vmware/terraform-provider-vcfa/pull/75))
- **New Data Source:** `vcfa_content_library_item` to read Content Library Items ([#13](https://github.com/vmware/terraform-provider-vcfa/pull/13), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66))
- **New Resource:** `vcfa_org_oidc` to manage OpenID Connect settings for an Organization ([#14](https://github.com/vmware/terraform-provider-vcfa/pull/14))
- **New Data Source:** `vcfa_org_oidc` to read OpenID Connect settings from an Organization ([#14](https://github.com/vmware/terraform-provider-vcfa/pull/14))
- **New Resource:** `vcfa_org_networking` to manage Org Networking Settings ([#15](https://github.com/vmware/terraform-provider-vcfa/pull/15))
- **New Data Source:** `vcfa_org_networking` to read Org Networking Settings ([#15](https://github.com/vmware/terraform-provider-vcfa/pull/15))
- **New Data Source:** `vcfa_right` to read existing Rights ([#16](https://github.com/vmware/terraform-provider-vcfa/pull/16))
- **New Resource:** `vcfa_rights_bundle` to manage Rights Bundles ([#17](https://github.com/vmware/terraform-provider-vcfa/pull/17), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66))
- **New Data Source:** `vcfa_rights_bundle` to read existing Rights Bundles ([#17](https://github.com/vmware/terraform-provider-vcfa/pull/17), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66))
- **New Resource:** `vcfa_org_regional_networking` to manage Org Regional Networking Settings ([#18](https://github.com/vmware/terraform-provider-vcfa/pull/18), [#70](https://github.com/vmware/terraform-provider-vcfa/pull/70))
- **New Data Source:** `vcfa_org_regional_networking` to read Org Regional Networking Settings ([#18](https://github.com/vmware/terraform-provider-vcfa/pull/18))
- **New Resource:** `vcfa_role` to manage Roles ([#19](https://github.com/vmware/terraform-provider-vcfa/pull/19))
- **New Data Source:** `vcfa_role` to read existing Roles ([#19](https://github.com/vmware/terraform-provider-vcfa/pull/19))
- **New Resource:** `vcfa_org_regional_networking_vpc_qos` to manage Org Regional Networking VPC QoS ([#20](https://github.com/vmware/terraform-provider-vcfa/pull/20))
- **New Data Source:** `vcfa_org_regional_networking_vpc_qos` to read Org Regional Networking VPC QoS ([#20](https://github.com/vmware/terraform-provider-vcfa/pull/20))
- **New Resource:** `vcfa_global_role` to manage Global Roles ([#21](https://github.com/vmware/terraform-provider-vcfa/pull/21), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66))
- **New Data Source:** `vcfa_global_role` to read existing Global Roles ([#21](https://github.com/vmware/terraform-provider-vcfa/pull/21), [#66](https://github.com/vmware/terraform-provider-vcfa/pull/66))
- **New Resource:** `vcfa_api_token` to manage API Tokens ([#22](https://github.com/vmware/terraform-provider-vcfa/pull/22), [#62](https://github.com/vmware/terraform-provider-vcfa/pull/62))
- **New Resource:** `vcfa_certificate` to manage Certificates from the Certificates Library ([#23](https://github.com/vmware/terraform-provider-vcfa/pull/23), [#62](https://github.com/vmware/terraform-provider-vcfa/pull/62))
- **New Data Source:** `vcfa_certificate` to read Certificates from the Certificates Library ([#23](https://github.com/vmware/terraform-provider-vcfa/pull/23), [#62](https://github.com/vmware/terraform-provider-vcfa/pull/62))
- **New Resource:** `vcfa_org_local_user` to manage Org Local users ([#25](https://github.com/vmware/terraform-provider-vcfa/pull/25), [#40](https://github.com/vmware/terraform-provider-vcfa/pull/40))
- **New Data Source:** `vcfa_org_local_user` to read Org Local users ([#25](https://github.com/vmware/terraform-provider-vcfa/pull/25), [#40](https://github.com/vmware/terraform-provider-vcfa/pull/40))
- **New Resource:** `vcfa_org_ldap` to manage LDAP settings of Organizations ([#26](https://github.com/vmware/terraform-provider-vcfa/pull/26), [#28](https://github.com/vmware/terraform-provider-vcfa/pull/28), [#50](https://github.com/vmware/terraform-provider-vcfa/pull/50), [#61](https://github.com/vmware/terraform-provider-vcfa/pull/61), [#65](https://github.com/vmware/terraform-provider-vcfa/pull/65))
- **New Data Source:** `vcfa_org_ldap` to read LDAP settings of Organizations ([#26](https://github.com/vmware/terraform-provider-vcfa/pull/26), [#28](https://github.com/vmware/terraform-provider-vcfa/pull/28), [#50](https://github.com/vmware/terraform-provider-vcfa/pull/50), [#61](https://github.com/vmware/terraform-provider-vcfa/pull/61), [#65](https://github.com/vmware/terraform-provider-vcfa/pull/65))
- **New Resource:** `vcfa_provider_ldap` to manage global Provider LDAP settings ([#28](https://github.com/vmware/terraform-provider-vcfa/pull/28), [#65](https://github.com/vmware/terraform-provider-vcfa/pull/65))
- **New Data Source:** `vcfa_provider_ldap` to read global Provider LDAP settings ([#28](https://github.com/vmware/terraform-provider-vcfa/pull/28), [#65](https://github.com/vmware/terraform-provider-vcfa/pull/65))
- **New Data Source:** `vcfa_region_vm_class` to read Region VM Classes ([#31](https://github.com/vmware/terraform-provider-vcfa/pull/31))
- **New Resource:** `vcfa_supervisor_namespace` to manage Supervisor Namespaces ([#35](https://github.com/vmware/terraform-provider-vcfa/pull/35), [#58](https://github.com/vmware/terraform-provider-vcfa/pull/58), [#59](https://github.com/vmware/terraform-provider-vcfa/pull/59), [#80](https://github.com/vmware/terraform-provider-vcfa/pull/80), [#81](https://github.com/vmware/terraform-provider-vcfa/pull/81), [#98](https://github.com/vmware/terraform-provider-vcfa/pull/98))
- **New Data Source:** `vcfa_supervisor_namespace` to read Supervisor Namespaces ([#35](https://github.com/vmware/terraform-provider-vcfa/pull/35), [#58](https://github.com/vmware/terraform-provider-vcfa/pull/58), [#59](https://github.com/vmware/terraform-provider-vcfa/pull/59))
- **New Data Source:** `vcfa_kubeconfig` to get Kubeconfig ([#35](https://github.com/vmware/terraform-provider-vcfa/pull/35), [#59](https://github.com/vmware/terraform-provider-vcfa/pull/59))
- **New Resource:** `vcfa_org_settings` to manage Organization general settings ([#50](https://github.com/vmware/terraform-provider-vcfa/pull/50), [#61](https://github.com/vmware/terraform-provider-vcfa/pull/61))
- **New Data Source:** `vcfa_org_settings` to read Organization general settings ([#50](https://github.com/vmware/terraform-provider-vcfa/pull/50), [#61](https://github.com/vmware/terraform-provider-vcfa/pull/61))

### IMPROVEMENTS

- The `url` field in `provider` configuration does not require specifying API endpoint `/api/`
  ([#47](https://github.com/vmware/terraform-provider-vcfa/pull/47))

### NOTES

- Initialize the VCFA Provider repository ([#1](https://github.com/vmware/terraform-provider-vcfa/pull/1))
- Add generic resource management functionality in `resource_generic_crud.go` ([#2](https://github.com/vmware/terraform-provider-vcfa/pull/2))
- Add leftover removal mechanism that makes `make cleanup` work ([#2](https://github.com/vmware/terraform-provider-vcfa/pull/2))
- Use only API v40+ ([#6](https://github.com/vmware/terraform-provider-vcfa/pull/6))
- Add GitHub Action to detect replaced go-vcloud-director SDK in `go.mod` in PRs ([#27](https://github.com/vmware/terraform-provider-vcfa/pull/27))
- Use `ClientContainer` to store provider SDK Clients, including `govcd.VCDClient` ([#29](https://github.com/vmware/terraform-provider-vcfa/pull/29))
- Bump github.com/hashicorp/yamux from v0.1.1 to v0.1.2 ([#30](https://github.com/vmware/terraform-provider-vcfa/pull/30))
- Add scripts to run Binary tests, that run HCL files generated by the regular Acceptance tests with a real
  Terraform executable ([#48](https://github.com/vmware/terraform-provider-vcfa/pull/48))
- Artifact cleanup will be attempted immediately after a failed test ([#52](https://github.com/vmware/terraform-provider-vcfa/pull/52))
- Add [`TROUBLESHOOTING.md`](./TROUBLESHOOTING.md) document to help diagnosing common issues ([#57](https://github.com/vmware/terraform-provider-vcfa/pull/57), [#63](https://github.com/vmware/terraform-provider-vcfa/pull/63))
- Reuse vCenter and NSX Manager in tests to performance and reliability ([#60](https://github.com/vmware/terraform-provider-vcfa/pull/60))
- New guide to [Import resources](https://registry.terraform.io/providers/vmware/vcfa/latest/docs/guides/importing_resources) ([#62](https://github.com/vmware/terraform-provider-vcfa/pull/62))
- New guide for [Roles management](https://registry.terraform.io/providers/vmware/vcfa/latest/docs/guides/roles_management) ([#71](https://github.com/vmware/terraform-provider-vcfa/pull/71))
- Add Broadcom licenses to .go files and `make licensecheck` command that is run in GitHub actions
  ([#73](https://github.com/vmware/terraform-provider-vcfa/pull/73))
- Change documentation to follow the modern Terraform layout for providers ([#74](https://github.com/vmware/terraform-provider-vcfa/pull/74), [#79](https://github.com/vmware/terraform-provider-vcfa/pull/79))
- Add examples of how to use the VMware Cloud Foundation Automation Terraform Provider, that can be found [here](examples) ([#81](https://github.com/vmware/terraform-provider-vcfa/pull/81))
- Migrate tests to `hashicorp/terraform-plugin-testing` ([#83](https://github.com/vmware/terraform-provider-vcfa/pull/83))
- Consume go-vcloud-director v3.0.0 (SDK this provider uses for low level access to Tenant Manager side of VCFA) ([#100](https://github.com/vmware/terraform-provider-vcfa/pull/100))
