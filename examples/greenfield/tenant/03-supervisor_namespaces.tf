# Create a Supervisor namespace

# https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest
resource "kubernetes_manifest" "example" {
  manifest = {
    "apiVersion" = "project.cci.vmware.com/v1alpha2"
    "kind"       = "Project"
    "metadata" = {
      "name" = "tf-tenant-example-project"
    }
    "spec" = {
      "description" = "Created by Terraform VCFA Provider"
    }
  }
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/supervisor_namespace
resource "vcfa_supervisor_namespace" "example" {
  depends_on = [
    kubernetes_manifest.example
  ]

  name_prefix  = "tf-tenant-example-supervisor-ns"
  project_name = "tf-tenant-example-project"
  class_name   = "small"
  description  = "Created by Terraform VCFA Provider"
  region_name  = local.region_name
  vpc_name     = format("%s-%s", local.region_name, "Default-VPC")

  storage_classes_initial_class_config_overrides {
    limit = "200Mi"
    name  = var.storage_class
  }

  zones_initial_class_config_overrides {
    cpu_limit          = "100M"
    cpu_reservation    = "1M"
    memory_limit       = "200Mi"
    memory_reservation = "2Mi"
    name               = var.supervisor_zone
  }
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/kubeconfig
data "vcfa_kubeconfig" "example_supervisor_namespace" {
  project_name              = kubernetes_manifest.example.manifest.metadata.name
  supervisor_namespace_name = vcfa_supervisor_namespace.example.name
}
