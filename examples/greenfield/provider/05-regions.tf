# Create a Region

data "vcfa_supervisor" "demo" {
  name       = var.supervisor_name
  vcenter_id = vcfa_vcenter.demo.id
}

data "vcfa_supervisor_zone" "demo" {
  supervisor_id = data.vcfa_supervisor.demo.id
  name          = var.supervisor_zone_name
}

resource "vcfa_region" "demo" {
  name                 = format("%s-%s", "tf-demo-region", substr(md5(var.url), 0, 4))
  description          = "Created by Terraform VCFA Provider"
  nsx_manager_id       = vcfa_nsx_manager.demo.id
  supervisor_ids       = [data.vcfa_supervisor.demo.id]
  storage_policy_names = var.vcenter_storage_policy_names[*]
}