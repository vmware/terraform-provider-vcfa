// © Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"os"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/util"
)

// isSystem returns true if the given Organization is System (Provider)
func isSystem(adminOrg *govcd.TmOrg) bool {
	return strings.EqualFold(adminOrg.TmOrg.Name, "system")
}

// Returns a valid Tenant Context if the Organization identified by the given ID is valid and exists.
// Otherwise, it returns either an empty tenant context, or an error if the Organization does not exist or is invalid.
func getTenantContextFromOrgId(tmClient *VCDClient, orgId string) (*govcd.TenantContext, error) {
	if orgId == "" {
		return &govcd.TenantContext{}, nil
	}
	org, err := tmClient.GetTmOrgById(orgId)
	if err != nil {
		return nil, err
	}
	return &govcd.TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	}, nil
}

// safeClose closes a file and logs the error, if any. This can be used instead of file.Close()
func safeClose(file *os.File) {
	if err := file.Close(); err != nil {
		util.Logger.Printf("Error closing file: %s\n", err)
	}
}
