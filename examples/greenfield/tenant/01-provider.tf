# Configure the VMware Cloud Foundation Automation Terraform Provider

terraform {
  required_providers {
    vcfa = {
      source  = "vmware/vcfa"
      version = "~> 1.0.0"
    }

    kubernetes = {
      source = "hashicorp/kubernetes"
    }
  }
}

provider "vcfa" {
  user                 = var.username
  password             = var.password
  url                  = var.url
  org                  = var.org
  allow_unverified_ssl = "true"
  logging              = true
}

locals {
  region_name = format("%s-%s", var.region_name, substr(md5(var.url), 0, 4))
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/org
data "vcfa_org" "example" {
  name = var.org
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/region
data "vcfa_region" "example" {
  name = local.region_name
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/kubeconfig
data "vcfa_kubeconfig" "kube_config" {}

# Initialize kubernetes provider leveraging kubeconfig data. Read more at:
# https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs
provider "kubernetes" {
  host     = data.vcfa_kubeconfig.kube_config.host
  insecure = data.vcfa_kubeconfig.kube_config.insecure_skip_tls_verify
  token    = data.vcfa_kubeconfig.kube_config.token
}
