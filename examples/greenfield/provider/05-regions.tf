# Create a Region

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/supervisor
data "vcfa_supervisor" "example" {
  name       = var.supervisor_name
  vcenter_id = vcfa_vcenter.example.id
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/data-sources/supervisor_zone
data "vcfa_supervisor_zone" "example" {
  supervisor_id = data.vcfa_supervisor.example.id
  name          = var.supervisor_zone_name
}

# https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/region
resource "vcfa_region" "example" {
  name                 = format("%s-%s", "tf-example-region", substr(md5(var.url), 0, 4))
  description          = "Created by Terraform VCFA Provider"
  nsx_manager_id       = vcfa_nsx_manager.example.id
  supervisor_ids       = [data.vcfa_supervisor.example.id]
  storage_policy_names = var.vcenter_storage_policy_names[*]
}
