terraform {
  required_providers {
    vcfa = {
      source  = "vmware/vcfa"
      version = "~> 1.0.0"
    }
  }
}

provider "vcfa" {
  user     = var.username
  password = var.password

  url                  = var.url
  org                  = "System"
  allow_unverified_ssl = "true"
  logging              = true
}
