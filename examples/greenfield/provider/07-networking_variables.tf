variable "tier0_gateway_name" {
  type        = string
  description = "Name of NSX Tier-0 Gateway"
}

variable "nsx_edge_cluster_name" {
  type        = string
  description = "Name of NSX Edge Cluster attached to Tier-0 Gateway"
}