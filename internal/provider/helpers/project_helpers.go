// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

func GetProject(tmClient *vcfa.VCDClient, projectName string) (ccitypes.Project, error) {
	var project ccitypes.Project

	projectURL, err := tmClient.VCDClient.Client.GetEntityUrl(fmt.Sprintf("%s/%s", ccitypes.ProjectsURL, projectName))
	if err != nil {
		return project, fmt.Errorf("error getting project URL: %s", err)
	}

	if err := tmClient.VCDClient.Client.GetEntity(projectURL, nil, &project, nil); err != nil {
		if govcd.ContainsNotFound(err) {
			return project, fmt.Errorf("project %s not found", projectName)
		}
		return project, fmt.Errorf("error getting project %s: %s", projectName, err.Error())
	}

	return project, nil
}
