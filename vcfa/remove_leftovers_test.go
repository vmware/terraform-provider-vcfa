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
	// {Type: "vcfa_xxx", Name: "test_entity", Comment: "loaded with provisioning"},
}

// alsoDelete contains a list of entities that should be removed , in addition to the ones
// found by name matching
// Add to this list if you ever get an entity left behind by a test
var alsoDelete = entityList{
	// {Type: "vcfa_xxx", Name: "custom-name", Comment: "manually created"},
}

// isTest is a regular expression that tells if an entity needs to be deleted
var isTest = regexp.MustCompile(`^[Tt]est`)

// alwaysShow lists the resources that will always be shown
var alwaysShow = []string{
	"vcfa_vcenter",
}

func removeLeftovers(govcdClient *govcd.VCDClient, verbose bool) error {
	if verbose {
		fmt.Printf("Start leftovers removal\n")
	}

	// --------------------------------------------------------------
	// vCenters
	// --------------------------------------------------------------
	if govcdClient.Client.IsSysAdmin {
		allVcs, err := govcdClient.GetAllVCenters(nil)
		if err != nil {
			return fmt.Errorf("error retrieving provider vCenters: %s", err)
		}
		for _, vc := range allVcs {
			toBeDeleted := shouldDeleteEntity(alsoDelete, doNotDelete, vc.VSphereVCenter.Name, "vcfa_vcenter", 0, verbose)
			if toBeDeleted {
				fmt.Printf("\t REMOVING vCenter %s\n", vc.VSphereVCenter.Name)
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
