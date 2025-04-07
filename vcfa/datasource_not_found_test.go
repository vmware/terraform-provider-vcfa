/*
 * // © Broadcom. All Rights Reserved.
 * // The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
 * // SPDX-License-Identifier: MPL-2.0
 */

//go:build ALL || functional

package vcfa

import (
	"regexp"
	"testing"

	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccDataSourceNotFound is using Go sub-tests to ensure that "read" methods for all (current and future) data
// sources defined in this provider always return error and substring 'govcd.ErrorEntityNotFound' in it when an object
// is not found.
func TestAccDataSourceNotFound(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)

	// Exit the test early
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Setup temporary client to evaluate versions and conditionally skip tests
	tmClient := createTemporaryVCFAConnection(false)

	// Run a sub-test for each of data source defined in provider
	for _, dataSource := range Provider().DataSources() {
		t.Run(dataSource.Name, testSpecificDataSourceNotFound(dataSource.Name, tmClient))
	}
}

func testSpecificDataSourceNotFound(dataSourceName string, tmClient *VCDClient) func(*testing.T) {
	return func(t *testing.T) {
		type skipAlways struct {
			dataSourceName string
			reason         string
		}

		skipAlwaysSlice := []skipAlways{
			{
				dataSourceName: "vcfa_version",
				reason:         "Data source vcfa_version always returns data, it is not possible to get ENF",
			},
			{
				dataSourceName: "vcfa_kubeconfig",
				reason:         "Data source vcfa_kubeconfig always returns data, it is not possible to get ENF",
			},
			{
				// TODO: TM: Retrieving non-existent Supervisor by ID returns 400 and not ENF
				dataSourceName: "vcfa_supervisor_zone",
				reason:         "TODO: TM: Retrieving non-existent Supervisor by ID returns 400 and not ENF",
			},
			{
				dataSourceName: "vcfa_provider_ldap",
				reason:         "Data source vcfa_provider_ldap always returns data, it is not possible to get ENF",
			},
			{
				dataSourceName: "vcfa_supervisor_namespace",
				reason:         "Data source vcfa_supervisor_namespace requires different auth mechanism",
			},
		}
		for _, skip := range skipAlwaysSlice {
			if dataSourceName == skip.dataSourceName {
				t.Skipf("Skipping: %s", skip.reason)
			}
		}

		// Skip subtest based on versions
		type skipOnVersion struct {
			skipVersionConstraint string
			datasourceName        string
		}

		skipOnVersionsVersionsOlderThan := []skipOnVersion{}

		for _, constraintSkip := range skipOnVersionsVersionsOlderThan {
			if dataSourceName == constraintSkip.datasourceName && tmClient.Client.APIVCDMaxVersionIs(constraintSkip.skipVersionConstraint) {
				t.Skipf("This test does not work on API versions %s", constraintSkip.skipVersionConstraint)
			}
		}

		// Skip sub-test if conditions are not met
		dataSourcesRequiringSysAdmin := []string{
			"vcfa_org",
			"vcfa_region",
			"vcfa_supervisor",
			"vcfa_supervisor_zone",
			"vcfa_vcenter",
			"vcfa_ip_space",
			"vcfa_region_zone",
			"vcfa_org_region_quota",
			"vcfa_region_vm_class",
			"vcfa_tier0_gateway",
			"vcfa_content_library",
			"vcfa_content_library_item",
		}

		if contains(dataSourcesRequiringSysAdmin, dataSourceName) && !usingSysAdmin() {
			t.Skip(`Works only with system admin privileges`)
		}

		// Get list of mandatory fields in schema for a particular data source
		mandatoryFields := getMandatoryDataSourceSchemaFields(dataSourceName)
		addedParams := addMandatoryParams(dataSourceName, mandatoryFields, t, tmClient)

		var params = StringMap{
			"DataSourceName":  dataSourceName,
			"MandatoryFields": addedParams,
		}

		params["FuncName"] = "NotFoundDataSource-" + dataSourceName
		// Adding skip directive as running these tests in binary test mode add no value
		binaryTestSkipText := "# skip-binary-test: data source not found test only works in acceptance tests\n"
		configText := templateFill(binaryTestSkipText+testAccUnavailableDataSource, params)

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config:      configText,
					ExpectError: regexp.MustCompile(`.*` + regexp.QuoteMeta(govcd.ErrorEntityNotFound.Error()) + `.*`),
				},
			},
		})
	}
}

const testAccUnavailableDataSource = `
data "{{.DataSourceName}}" "not-existing" {
  {{.MandatoryFields}}
}
`

// getMandatoryDataSourceSchemaFields checks schema definitions for data sources and return slice of mandatory fields
func getMandatoryDataSourceSchemaFields(dataSourceName string) []string {
	var mandatoryFields []string
	schema := globalDataSourceMap[dataSourceName]
	for fieldName, fieldSchema := range schema.Schema {
		if fieldSchema.Required || (len(fieldSchema.ExactlyOneOf) > 0 && fieldSchema.ExactlyOneOf[0] == fieldName) {
			mandatoryFields = append(mandatoryFields, fieldName)
		}
	}
	return mandatoryFields
}

func addMandatoryParams(dataSourceName string, mandatoryFields []string, t *testing.T, tmClient *VCDClient) string {
	var templateFields string
	for fieldIndex := range mandatoryFields {
		switch mandatoryFields[fieldIndex] {
		case "name":
			templateFields = templateFields + `name = "does-not-exist"` + "\n"
		case "project_name":
			templateFields = templateFields + `project_name = "does-not-exist"` + "\n"
		case "supervisor_id":
			templateFields = templateFields + `supervisor_id = "urn:vcloud:supervisor:12345678-1234-1234-1234-123456789012"` + "\n"
		case "vcenter_id":
			templateFields = templateFields + `vcenter_id = "urn:vcloud:vimserver:12345678-1234-1234-1234-123456789012"` + "\n"
		case "region_id":
			templateFields = templateFields + `region_id = "urn:vcloud:region:12345678-1234-1234-1234-123456789012"` + "\n"
		case "org_id":
			templateFields = templateFields + `org_id = "urn:vcloud:org:12345678-1234-1234-1234-123456789012"` + "\n"
		case "edge_cluster_id":
			templateFields = templateFields + `edge_cluster_id = "urn:vcloud:edgeCluster:12345678-1234-1234-1234-123456789012"` + "\n"
		case "content_library_id":
			templateFields = templateFields + `content_library_id = "urn:vcloud:contentLibrary:12345678-1234-1234-1234-123456789012"` + "\n"
		case "alias": // vcfa_certificate
			templateFields = templateFields + `alias = "non-existent-certificate"` + "\n"
		case "username": // vcfa_org_local_user
			templateFields = templateFields + `username = "non-existent-local-user"` + "\n"
		case "org_regional_networking_id":
			templateFields = templateFields + `org_regional_networking_id = "urn:vcloud:regionalNetworkingSetting:12345678-1234-1234-1234-123456789012"` + "\n"
		}
	}

	return templateFields
}
