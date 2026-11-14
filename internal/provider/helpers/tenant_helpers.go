// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

// GetTenantContext returns the tenant context of the Org with the given ID, which includes the Org name.
func GetTenantContext(tmClient *govcd.VCDClient, orgID string) (*govcd.TenantContext, error) {
	org, err := tmClient.GetTmOrgById(orgID)
	if err != nil {
		return nil, err
	}
	return &govcd.TenantContext{OrgId: org.TmOrg.ID, OrgName: org.TmOrg.Name}, nil
}
