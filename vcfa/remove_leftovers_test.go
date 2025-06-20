// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/govcd"
)

// This file contains routines that clean up the test suite after failed tests

// entityDef is the definition of an entity (to be either deleted or kept)
// with an optional comment
type entityDef struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Comment    string `json:"comment,omitempty"`
	NameRegexp *regexp.Regexp
}

// entityList is a collection of entityDef
type entityList []entityDef

// doNotDelete contains a list of entities that should not be deleted,
// despite having a name that starts with `Test` or `test`
var doNotDelete = entityList{
	{Type: "vcfa_org", Name: "System", Comment: "Built-in admin Org"},
	{Type: "vcfa_org", Name: "tenant1", Comment: "tenant loaded with provisioning"},
	{Type: "vcfa_org", Name: "system-classic-tenant", Comment: "tenant loaded with provisioning"},
	{Type: "vcfa_org", Name: "tenant1classic", Comment: "classic tenant loaded with provisioning"},
}

// alsoDelete contains a list of entities that should be removed , in addition to the ones
// found by name matching
// Add to this list if you ever get an entity left behind by a test
var alsoDelete = entityList{}

// isTest is a regular expression that tells if an entity needs to be deleted
var isTest = regexp.MustCompile(`^[Tt]est`)

// alwaysShow lists the resources that will always be shown
var alwaysShow = []string{
	"vcfa_vcenter",
	"vcfa_org",
	"vcfa_ip_space",
	"vcfa_provider_gateway",
	"vcfa_org_regional_networking",
	"vcfa_edge_cluster_qos",
	"vcfa_content_library",
	"vcfa_region",
	"vcfa_nsx_manager",
}

func removeLeftovers(tmClient *govcd.VCDClient, verbose, isFinalCleanup bool) error {
	if verbose {
		fmt.Printf("Start leftovers removal\n")
	}

	// TODO - improve the skip to avoid cleaning up
	if !tmClient.Client.IsSysAdmin {
		fmt.Println("Skipping leftover removal for Org user mode")
		return nil
	}

	// --------------------------------------------------------------
	// Remove all children from existing Tenants
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin {
		orgs, err := tmClient.GetAllTmOrgs(nil)
		if err != nil {
			return fmt.Errorf("error retrieving Organizations: %s", err)
		}
		for _, org := range orgs {
			if org.TmOrg.IsClassicTenant {
				continue
			}
			// --------------------------------------------------------------
			// Tenant Content Libraries
			// --------------------------------------------------------------
			var cls []*govcd.ContentLibrary
			cls, err = org.GetAllContentLibraries(nil)
			if err != nil {
				return fmt.Errorf("error retrieving Tenant '%s' Content Libraries: %s", org.TmOrg.Name, err)
			}
			err = removeLeftoverContentLibraries(cls, 1, verbose)
			if err != nil {
				return err
			}
		}
	}

	// --------------------------------------------------------------
	// Org Regional Network Configuration
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin {
		all, err := tmClient.GetAllTmRegionalNetworkingSettings(nil)
		if err != nil {
			return fmt.Errorf("error retrieving All Regional Networking Settings: %s", err)
		}
		for _, one := range all {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, one.TmRegionalNetworkingSetting.Name, "vcfa_org_regional_networking", 3, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING All %s Settings %s\n", labelVcfaRegionalNetworkingSetting, one.TmRegionalNetworkingSetting.Name)
				err := one.Delete()
				if err != nil {
					return fmt.Errorf("error deleting %s Settings '%s': %s", labelVcfaRegionalNetworkingSetting, one.TmRegionalNetworkingSetting.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// Edge Cluster QoS (Edge Clusters themselves are read-only)
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin {
		allEcs, err := tmClient.GetAllTmEdgeClusters(nil)
		if err != nil {
			return fmt.Errorf("error retrieving Edge Clusters: %s", err)
		}
		for _, ec := range allEcs {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, ec.TmEdgeCluster.Name, "vcfa_edge_cluster_qos", 2, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING Edge Cluster QoS Settings %s\n", ec.TmEdgeCluster.Name)
				err := ec.Delete()
				if err != nil {
					return fmt.Errorf("error deleting %s '%s': %s", labelVcfaEdgeClusterQos, ec.TmEdgeCluster.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// Provider Content Libraries
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin {
		cls, err := tmClient.GetAllContentLibraries(nil, nil)
		if err != nil {
			return fmt.Errorf("error retrieving Provider Content Libraries: %s", err)
		}
		for _, cl := range cls {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, cl.ContentLibrary.Name, "vcfa_content_library", 0, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING %s %s\n", labelVcfaContentLibrary, cl.ContentLibrary.Name)
				err := cl.Delete(true, true)
				if err != nil {
					return fmt.Errorf("error deleting %s '%s': %s", labelVcfaContentLibrary, cl.ContentLibrary.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// Provider Gateways
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin {
		pgs, err := tmClient.GetAllTmProviderGateways(nil)
		if err != nil {
			return fmt.Errorf("error retrieving IP Spaces: %s", err)
		}
		for _, pg := range pgs {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, pg.TmProviderGateway.Name, "vcfa_provider_gateway", 2, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING %s %s\n", labelVcfaProviderGateway, pg.TmProviderGateway.Name)
				err := pg.Delete()
				if err != nil {
					return fmt.Errorf("error deleting %s '%s': %s", labelVcfaProviderGateway, pg.TmProviderGateway.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// IP Spaces
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin {
		ipSpaces, err := tmClient.GetAllTmIpSpaces(nil)
		if err != nil {
			return fmt.Errorf("error retrieving IP Spaces: %s", err)
		}
		for _, ipSp := range ipSpaces {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, ipSp.TmIpSpace.Name, "vcfa_ip_space", 2, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING %s %s\n", labelVcfaIpSpace, ipSp.TmIpSpace.Name)
				err := ipSp.Delete()
				if err != nil {
					return fmt.Errorf("error deleting %s '%s': %s", labelVcfaIpSpace, ipSp.TmIpSpace.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// Region Quotas
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin {
		rqs, err := tmClient.GetAllRegionQuotas(nil)
		if err != nil {
			return fmt.Errorf("error retrieving Region Quotas: %s", err)
		}
		for _, rq := range rqs {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, rq.TmVdc.Name, "vcfa_org_region_quota", 2, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING %s %s\n", labelVcfaOrgRegionQuota, rq.TmVdc.Name)
				err := rq.Delete()
				if err != nil {
					return fmt.Errorf("error deleting %s '%s': %s", labelVcfaOrgRegionQuota, rq.TmVdc.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// Regions
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin {
		regions, err := tmClient.GetAllRegions(nil)
		if err != nil {
			return fmt.Errorf("error retrieving Regions: %s", err)
		}
		for _, region := range regions {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, region.Region.Name, "vcfa_region", 1, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING %s %s\n", labelVcfaRegion, region.Region.Name)
				err := region.Delete()
				if err != nil {
					return fmt.Errorf("error deleting %s '%s': %s", labelVcfaRegion, region.Region.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// Organizations
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin {
		orgs, err := tmClient.GetAllTmOrgs(nil)
		if err != nil {
			return fmt.Errorf("error retrieving Organizations: %s", err)
		}
		for _, org := range orgs {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, org.TmOrg.Name, "vcfa_org", 0, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING Organization %s\n", org.TmOrg.Name)
				err = org.Disable()
				if err != nil {
					return fmt.Errorf("error disabling %s '%s': %s", labelVcfaOrg, org.TmOrg.Name, err)
				}
				err := org.Delete()
				if err != nil {
					return fmt.Errorf("error deleting %s '%s': %s", labelVcfaOrg, org.TmOrg.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// NSX Managers
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin && (isFinalCleanup || vcfaSkipPriorityTests) {
		allNsxManagers, err := tmClient.GetAllNsxtManagersOpenApi(nil)
		if err != nil {
			return fmt.Errorf("error retrieving provider NSX Managers: %s", err)
		}
		for _, m := range allNsxManagers {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, m.NsxtManagerOpenApi.Name, "vcfa_nsx_manager", 0, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING %s %s\n", labelVcfaNsxManager, m.NsxtManagerOpenApi.Name)
				err := m.Delete()
				if err != nil {
					return fmt.Errorf("error deleting %s '%s': %s", labelVcfaNsxManager, m.NsxtManagerOpenApi.Name, err)
				}
			}
		}
	}

	// --------------------------------------------------------------
	// vCenters
	// --------------------------------------------------------------
	if tmClient.Client.IsSysAdmin && (isFinalCleanup || vcfaSkipPriorityTests) {
		allVcs, err := tmClient.GetAllVCenters(nil)
		if err != nil {
			return fmt.Errorf("error retrieving provider vCenters: %s", err)
		}
		for _, vc := range allVcs {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, vc.VSphereVCenter.Name, "vcfa_vcenter", 0, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING %s %s\n", labelVcfaVirtualCenter, vc.VSphereVCenter.Name)
				err = vc.Disable()
				if err != nil {
					return fmt.Errorf("error disabling %s '%s': %s", labelVcfaVirtualCenter, vc.VSphereVCenter.Name, err)
				}
				err := vc.Delete()
				if err != nil {
					return fmt.Errorf("error deleting %s '%s': %s", labelVcfaVirtualCenter, vc.VSphereVCenter.Name, err)
				}
			}
		}
	}

	if verbose {
		fmt.Printf("End leftovers removal\n")
	}

	return nil
}

// shouldDeleteEntity checks whether a given entity is to be deleted, either by its name
// or by its inclusion in one of the entity lists
func shouldDeleteEntity(alsoDelete, doNotDelete entityList, name, entityType string, level int, verbose bool) bool {
	inclusion := ""
	exclusion := ""
	// 1. First requirement to be deleted: the entity name starts with 'Test' or 'test'
	toBeDeleted := isTest.MatchString(name)
	if inList(alsoDelete, name, entityType) {
		toBeDeleted = true
		// 2. If the entity was in the additional deletion list, regardless of the name,
		// it is marked for deletion, with a "+", indicating that it was selected for deletion because of the
		// deletion list
		inclusion = " +"
	}
	if inList(doNotDelete, name, entityType) {
		toBeDeleted = false
		// 3. If a file, normally marked for deletion, is found in the keep list,
		// its deletion status is revoked, and it is marked with a "-", indicating that it was excluded
		// for deletion because of the keep list
		exclusion = " -"
	}
	tabs := strings.Repeat("\t", level)
	format := tabs + "[%s] %s (%s%s%s)\n"
	deletionText := "DELETE"
	if !toBeDeleted {
		deletionText = "keep"
	}

	// 4. Show the entity. If it is to be deleted, it will always be shown
	if toBeDeleted || contains(alwaysShow, entityType) {
		if verbose {
			fmt.Printf(format, entityType, name, deletionText, inclusion, exclusion)
		}
	}
	return toBeDeleted
}

// inList shows whether a given entity is included in an entityList
func inList(list entityList, name, entityType string) bool {
	for _, element := range list {
		// Compare by names
		if element.Name == name && element.Type == entityType {
			return true
		}
		// Compare by possible regexp values
		if element.NameRegexp != nil && element.NameRegexp.MatchString(name) {
			return true
		}
	}
	return false
}

// removeLeftoverContentLibraries deletes the given Content Libraries
func removeLeftoverContentLibraries(cls []*govcd.ContentLibrary, level int, verbose bool) error {
	for _, cl := range cls {
		toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, cl.ContentLibrary.Name, "vcfa_content_library", level, verbose)
		if toBeDeleted {
			fmt.Printf("\t REMOVING %s %s\n", labelVcfaContentLibrary, cl.ContentLibrary.Name)
			err := cl.Delete(true, true)
			if err != nil {
				return fmt.Errorf("error deleting %s '%s': %s", labelVcfaContentLibrary, cl.ContentLibrary.Name, err)
			}
		}
	}
	return nil
}
