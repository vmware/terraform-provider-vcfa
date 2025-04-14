# Configure networking

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/ip_space
resource "vcfa_ip_space" "example" {
  count = 2

  name                          = "tf-example-ip-space-${count.index}"
  description                   = "IP Space ${count.index} - Created by Terraform VCFA Provider"
  region_id                     = vcfa_region.example.id
  external_scope                = "1${count.index}.12.0.0/30"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1
  internal_scope {
    name = format("%s-%s-1", substr(md5(var.url), 0, 8), count.index)
    cidr = "10.${count.index}.0.0/28"
  }
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/tier0_gateway
data "vcfa_tier0_gateway" "example" {
  name      = var.tier0_gateway_name
  region_id = vcfa_region.example.id
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/provider_gateway
resource "vcfa_provider_gateway" "example" {
  name             = "tf-example-provider-gateway"
  description      = "Created by Terraform VCFA Provider"
  region_id        = vcfa_region.example.id
  tier0_gateway_id = data.vcfa_tier0_gateway.example.id
  ip_space_ids     = [vcfa_ip_space.example[0].id]
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/edge_cluster
data "vcfa_edge_cluster" "example" {
  name             = var.nsx_edge_cluster_name
  region_id        = vcfa_region.example.id
  sync_before_read = true
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/edge_cluster_qos
resource "vcfa_edge_cluster_qos" "example" {
  edge_cluster_id = data.vcfa_edge_cluster.example.id

  egress_committed_bandwidth_mbps  = 1
  egress_burst_size_bytes          = 2
  ingress_committed_bandwidth_mbps = 3
  ingress_burst_size_bytes         = 4
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org_networking
resource "vcfa_org_networking" "example" {
  org_id   = vcfa_org.example.id
  log_name = substr(md5(var.url), 0, 6)
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org_regional_networking
resource "vcfa_org_regional_networking" "example" {
  name = "tf-example-regional-networking"
  # referencing vcfa_org_networking instead of vcfa_org to preserve correct order of actions
  org_id              = vcfa_org_networking.example.id
  provider_gateway_id = vcfa_provider_gateway.example.id
  region_id           = vcfa_region.example.id
  edge_cluster_id     = data.vcfa_edge_cluster.example.id
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org_regional_networking_vpc_qos
resource "vcfa_org_regional_networking_vpc_qos" "example" {
  org_regional_networking_id       = vcfa_org_regional_networking.example.id
  ingress_committed_bandwidth_mbps = 14
  ingress_burst_size_bytes         = 15
  egress_committed_bandwidth_mbps  = 16
  egress_burst_size_bytes          = 17
}
