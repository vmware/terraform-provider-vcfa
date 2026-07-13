// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package testutils

import (
	"fmt"
	"testing"

	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetupProject creates a VCFA Project using a direct go-vcloud-director connection and
// returns a cleanup function that deletes it. It mirrors the helper previously embedded in
// the supervisor namespace acceptance test.
func SetupProject(t *testing.T, projectName string) func() {
	t.Helper()
	client := NewClient(t)

	projectCfg := &ccitypes.Project{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.ProjectKind,
			APIVersion: ccitypes.ProjectAPI + "/" + ccitypes.ProjectVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: projectName,
		},
		Spec: ccitypes.ProjectSpec{
			Description: fmt.Sprintf("Terraform test project [%s]", projectName),
		},
	}

	newProjectAddr, err := client.Client.GetEntityUrl(ccitypes.ProjectsURL)
	if err != nil {
		t.Fatalf("error creating URL for new project: %s", err)
	}

	newProject := &ccitypes.Project{}
	if err := client.Client.PostEntity(newProjectAddr, nil, projectCfg, newProject, nil); err != nil {
		t.Fatalf("error creating project %s: %s", projectCfg.Name, err)
	}

	return func() {
		projectAddr, err := client.Client.GetEntityUrl(ccitypes.ProjectsURL, "/", projectCfg.Name)
		if err != nil {
			t.Fatalf("error getting Project url: %s", err)
		}
		if err := client.Client.DeleteEntity(projectAddr, nil, nil); err != nil {
			t.Fatalf("failed removing Project: %s", err)
		}
	}
}
