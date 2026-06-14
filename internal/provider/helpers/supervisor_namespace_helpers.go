// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

func GetSupervisorNamespaceEndpointURL(tmClient *vcfa.VCDClient, projectName string, supervisorNamespaceName string) (string, error) {
	if _, err := GetProject(tmClient, projectName); err != nil {
		return "", fmt.Errorf("error getting project %s: %s", projectName, err)
	}

	supervisorNamespaceURL, err := buildSupervisorNamespaceURL(tmClient, projectName, supervisorNamespaceName)
	if err != nil {
		return "", fmt.Errorf("error getting supervisor namespace endpoint URL: %s", err)
	}

	var supervisorNamespace ccitypes.SupervisorNamespace
	if err := tmClient.VCDClient.Client.GetEntity(supervisorNamespaceURL, nil, &supervisorNamespace, nil); err != nil {
		if govcd.ContainsNotFound(err) {
			return "", fmt.Errorf("supervisor namespace %s not found in project %s", supervisorNamespaceName, projectName)
		}
		return "", fmt.Errorf("error getting supervisor namespace %s in project %s: %s", supervisorNamespaceName, projectName, err)
	}

	readyStatus := false
	for _, condition := range supervisorNamespace.Status.Conditions {
		if strings.ToLower(condition.Type) == "ready" {
			if strings.ToLower(condition.Status) == "true" {
				readyStatus = true
			}
			break
		}
	}
	if !readyStatus {
		return "", fmt.Errorf("supervisor namespace %s in project %s is not in a ready status", supervisorNamespaceName, projectName)
	}
	if supervisorNamespace.Status.NamespaceEndpointURL == "" {
		return "", fmt.Errorf("unable to retrieve the endpoint URL for supervisor namespace %s in project %s", supervisorNamespaceName, projectName)
	}

	return supervisorNamespace.Status.NamespaceEndpointURL, nil
}

func buildSupervisorNamespaceURL(tmClient *vcfa.VCDClient, projectName string, supervisorNamespaceName string) (*url.URL, error) {
	supervisorNamespaceRawURL := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, projectName)
	if supervisorNamespaceName != "" {
		supervisorNamespaceRawURL = supervisorNamespaceRawURL + "/" + supervisorNamespaceName
	}

	return tmClient.VCDClient.Client.GetEntityUrl(supervisorNamespaceRawURL)
}
