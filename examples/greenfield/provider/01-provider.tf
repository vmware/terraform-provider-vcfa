# Configure the VMware Cloud Foundation Automation Terraform Provider

terraform {
  required_providers {
    vcfa = {
      source  = "vmware/vcfa"
      version = "~> 1.0.0"
    }
  }
}

provider "vcfa" {
  user                 = var.username
  password             = var.password
  url                  = var.url
  org                  = "System" # Login in the Provider (System) org
  allow_unverified_ssl = "true"
  logging              = true # Generates the log file for troubleshooting
}

# https://registry.terraform.io/providers/vmware/vcd/latest/docs/data-sources/org
data "vcfa_org" "system" {
  name = "System"
}
