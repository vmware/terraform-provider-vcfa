//go:build tm || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/vmware/terraform-provider-vcfa/internal/testutils"
	"github.com/vmware/terraform-provider-vcfa/internal/testutils/providertest"
)

// TestAccVcfaSupervisorNamespaceClass exercises the lifecycle of the
// vcfa_supervisor_namespace_class resource and data source against a live environment: create with
// every attribute, update, read with the data source, import, and a minimal configuration.
func TestAccVcfaSupervisorNamespaceClass(t *testing.T) {
	testutils.SkipIfShort(t)

	cfg := testutils.GetTestConfig(t)
	sysAdmin := testutils.UsingSysAdmin(cfg)
	if len(cfg.Tm.RegionVmClasses) < 2 {
		t.Skip("tm.regionVmClasses must list at least two VM classes")
	}

	params := testutils.StringMap{
		"Name":         strings.ReplaceAll(strings.ToLower(t.Name()), "_", "-"),
		"Description":  "Created by " + t.Name(),
		"StorageClass": cfg.Tm.StorageClass,
		"VmClass1":     cfg.Tm.RegionVmClasses[0],
		"VmClass2":     cfg.Tm.RegionVmClasses[1],
		"Library":      cfg.Tm.ContentLibrary,
		"Zone":         cfg.Tm.VcenterSupervisorZone,
		"StorageLimit": "100Gi",
		"CpuLimit":     "10",
		"Count":        "2",

		"Org":         strings.ReplaceAll(strings.ToLower(t.Name()), "_", "-"),
		"OrgUser":     "tflocal",
		"OrgPassword": "long-change-ME1",
		"VcfaUrl":     cfg.Provider.Url,
		"Insecure":    cfg.Provider.AllowInsecure,
		"Provider":    "provider = vcfatenant",
	}
	testutils.TestParamsNotEmpty(t, params)

	prerequisites, tenantProvider := testAccSupervisorNamespaceClassPrerequisites, testAccSupervisorNamespaceClassTenantProvider
	if !sysAdmin {
		prerequisites, tenantProvider = "", ""
		params["Provider"] = ""
	}

	configText0 := testutils.TemplateFill(t, prerequisites, params)

	params["FuncName"] = t.Name() + "-create"
	configText1 := testutils.TemplateFill(t, prerequisites+tenantProvider+testAccSupervisorNamespaceClassFull, params)

	params["FuncName"] = t.Name() + "-update"
	params["Description"] = "Updated by " + t.Name()
	params["StorageLimit"] = "200Gi"
	params["CpuLimit"] = "20"
	params["Count"] = "3"
	configText2 := testutils.TemplateFill(t, prerequisites+tenantProvider+testAccSupervisorNamespaceClassFull, params)

	params["FuncName"] = t.Name() + "-ds"
	configText3 := testutils.TemplateFill(t, prerequisites+tenantProvider+testAccSupervisorNamespaceClassFull+testAccSupervisorNamespaceClassData, params)

	params["FuncName"] = t.Name() + "-default-limits"
	configText5 := testutils.TemplateFill(t, prerequisites+tenantProvider+testAccSupervisorNamespaceClassDefaultLimits, params)

	params["FuncName"] = t.Name() + "-wildcards"
	configText6 := testutils.TemplateFill(t, prerequisites+tenantProvider+testAccSupervisorNamespaceClassWildcards, params)

	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)

	const res = "vcfa_supervisor_namespace_class.test"
	name := params["Name"].(string)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providertest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// The tenant provider logs in when Terraform plans, so it can only be used once the Org
			// and its local user exist: the first step creates just them.
			{
				SkipFunc: func() (bool, error) { return !sysAdmin, nil },
				Config:   configText0,
				Check:    resource.TestCheckResourceAttrSet("vcfa_org_local_user.test", "id"),
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "id", name),
					resource.TestCheckResourceAttr(res, "name", name),
					resource.TestCheckResourceAttr(res, "description", "Created by "+t.Name()),
					resource.TestCheckResourceAttr(res, "storage_classes.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "storage_classes.*", map[string]string{"name": cfg.Tm.StorageClass, "limit": "100Gi"}),
					resource.TestCheckResourceAttr(res, "vm_classes.#", "2"),
					resource.TestCheckTypeSetElemAttr(res, "vm_classes.*", cfg.Tm.RegionVmClasses[0]),
					resource.TestCheckResourceAttr(res, "content_sources.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "content_sources.*", map[string]string{"name": cfg.Tm.ContentLibrary, "type": "ContentLibrary"}),
					resource.TestCheckResourceAttr(res, "zones.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "zones.*", map[string]string{
						"name": cfg.Tm.VcenterSupervisorZone, "cpu_limit": "10", "vm_class_reservations.#": "1",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "description", "Updated by "+t.Name()),
					resource.TestCheckTypeSetElemNestedAttrs(res, "storage_classes.*", map[string]string{"name": cfg.Tm.StorageClass, "limit": "200Gi"}),
					resource.TestCheckTypeSetElemNestedAttrs(res, "zones.*", map[string]string{"cpu_limit": "20"}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.vcfa_supervisor_namespace_class.test", "id", res, "id"),
					resource.TestCheckResourceAttrPair("data.vcfa_supervisor_namespace_class.test", "description", res, "description"),
					resource.TestCheckResourceAttrPair("data.vcfa_supervisor_namespace_class.test", "storage_classes.#", res, "storage_classes.#"),
					resource.TestCheckResourceAttrPair("data.vcfa_supervisor_namespace_class.test", "vm_classes.#", res, "vm_classes.#"),
					resource.TestCheckResourceAttrPair("data.vcfa_supervisor_namespace_class.test", "zones.#", res, "zones.#"),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     name,
				ImportStateVerify: true,
			},
			// Zones are replaced by default limits.
			{
				Config: configText5,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "default_limits.cpu_limit", "20"),
					resource.TestCheckResourceAttr(res, "default_limits.memory_limit", "64Gi"),
					resource.TestCheckResourceAttr(res, "zones.#", "0"),
				),
			},
			// Wildcards select every storage class, VM class and zone, as in the default classes.
			{
				Config: configText6,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(res, "storage_classes.*", map[string]string{"name": "*", "limit": "100000Mi"}),
					resource.TestCheckTypeSetElemAttr(res, "vm_classes.*", "*"),
					resource.TestCheckTypeSetElemNestedAttrs(res, "zones.*", map[string]string{
						"name": "*", "cpu_limit": "10000M", "cpu_reservation": "0M", "memory_limit": "10000Mi", "memory_reservation": "0Mi",
					}),
					resource.TestCheckResourceAttr(res, "default_limits.%", "0"),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     name,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccSupervisorNamespaceClassFull = `
resource "vcfa_supervisor_namespace_class" "test" {
  {{.Provider}}
  name        = "{{.Name}}"
  description = "{{.Description}}"

  storage_classes = [{ name = "{{.StorageClass}}", limit = "{{.StorageLimit}}" }]
  vm_classes      = ["{{.VmClass1}}", "{{.VmClass2}}"]
  content_sources = [{ name = "{{.Library}}" }]

  zones = [{
    name               = "{{.Zone}}"
    cpu_limit          = "{{.CpuLimit}}"
    memory_limit       = "64Gi"
    cpu_reservation    = "4"
    memory_reservation = "32Gi"
    vm_class_reservations = [{ vm_class_name = "{{.VmClass1}}", count = {{.Count}} }]
  }]
}
`

const testAccSupervisorNamespaceClassDefaultLimits = `
resource "vcfa_supervisor_namespace_class" "test" {
  {{.Provider}}
  name            = "{{.Name}}"
  storage_classes = [{ name = "{{.StorageClass}}", limit = "{{.StorageLimit}}" }]
  vm_classes      = ["{{.VmClass1}}"]

  default_limits = {
    cpu_limit          = "{{.CpuLimit}}"
    memory_limit       = "64Gi"
    cpu_reservation    = "4"
    memory_reservation = "32Gi"
  }
}
`

const testAccSupervisorNamespaceClassWildcards = `
resource "vcfa_supervisor_namespace_class" "test" {
  {{.Provider}}
  name            = "{{.Name}}"
  storage_classes = [{ name = "*", limit = "100000Mi" }]
  vm_classes      = ["*"]

  zones = [{
    name               = "*"
    cpu_limit          = "10000M"
    cpu_reservation    = "0M"
    memory_limit       = "10000Mi"
    memory_reservation = "0Mi"
  }]
}
`

const testAccSupervisorNamespaceClassData = `
data "vcfa_supervisor_namespace_class" "test" {
  {{.Provider}}
  name = vcfa_supervisor_namespace_class.test.name
}
`

const testAccSupervisorNamespaceClassPrerequisites = `
resource "vcfa_org" "test" {
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "terraform test"
  is_enabled   = true
}

data "vcfa_role" "org-admin" {
  org_id = vcfa_org.test.id
  name   = "Organization Administrator"
}

resource "vcfa_org_local_user" "test" {
  org_id   = vcfa_org.test.id
  role_ids = [data.vcfa_role.org-admin.id]
  username = "{{.OrgUser}}"
  password = "{{.OrgPassword}}"
}
`

// Provider logged in as the local user of the Org, as Supervisor Namespace Classes are tenant objects.
const testAccSupervisorNamespaceClassTenantProvider = `
provider "vcfatenant" {
  user                 = "{{.OrgUser}}"
  password             = "{{.OrgPassword}}"
  auth_type            = "integrated"
  url                  = "{{.VcfaUrl}}"
  org                  = "{{.Org}}"
  allow_unverified_ssl = {{.Insecure}}
}
`
