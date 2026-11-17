// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfatypes

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	LabelProject            = "Project"
	LabelProjectRoleBinding = "Project Role Binding"
)

const (
	ProjectAPI      = "project.cci.vmware.com"
	ProjectKind     = "Project"
	ProjectVersion  = "v1alpha2"
	ProjectResource = "projects"

	ProjectRoleBindingAPI      = "authorization.cci.vmware.com"
	ProjectRoleBindingVersion  = "v1alpha1"
	ProjectRoleBindingKind     = "ProjectRoleBinding"
	ProjectRoleBindingResource = "projectrolebindings"

	ProjectRoleKind = "ProjectRole"
)

// GetProjectGVR returns the GroupVersionResource for Project. Projects are cluster-scoped and
// served by the base CCI endpoint.
func GetProjectGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ProjectAPI,
		Version:  ProjectVersion,
		Resource: ProjectResource,
	}
}

// GetProjectRoleBindingGVR returns the GroupVersionResource for ProjectRoleBinding. Role bindings
// are namespaced, and their namespace is the name of their Project.
func GetProjectRoleBindingGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ProjectRoleBindingAPI,
		Version:  ProjectRoleBindingVersion,
		Resource: ProjectRoleBindingResource,
	}
}

// Subject kinds accepted by a ProjectRoleBinding.
const (
	ProjectRoleBindingSubjectKindUser  = "User"
	ProjectRoleBindingSubjectKindGroup = "Group"
)

// Built-in ProjectRoles. They map to the Project Administrator, Project User, Project Advanced User
// and Project Auditor roles respectively.
const (
	ProjectRoleAdmin   = "admin"
	ProjectRoleEdit    = "edit"
	ProjectRoleEditAdv = "edit_adv"
	ProjectRoleView    = "view"
)

// Project is a logical grouping of resources and users.
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProjectSpec `json:"spec"`
}

// ProjectSpec defines the desired state of a Project.
type ProjectSpec struct {
	Description string `json:"description"`
}

// NewProject builds the Project to create or update. The backend rejects a uid, labels or
// annotations in the payload.
func NewProject(name, description string) Project {
	return Project{
		TypeMeta: metav1.TypeMeta{
			Kind:       ProjectKind,
			APIVersion: ProjectAPI + "/" + ProjectVersion,
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       ProjectSpec{Description: description},
	}
}

// ProjectRoleBinding grants a ProjectRole to one user or group within a Project. The backend
// accepts exactly one subject per binding, and one binding per subject and Project.
type ProjectRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Subjects []ProjectRoleBindingSubject `json:"subjects,omitempty"`
	RoleRef  ProjectRoleBindingRoleRef   `json:"roleRef"`
}

// ProjectRoleBindingSubject identifies the user or group being granted the role.
type ProjectRoleBindingSubject struct {
	Kind string `json:"kind"`
	Name string `json:"name,omitempty"`
}

// ProjectRoleBindingRoleRef references the ProjectRole being granted.
type ProjectRoleBindingRoleRef struct {
	APIGroup string `json:"apiGroup,omitempty"`
	Kind     string `json:"kind,omitempty"`
	Name     string `json:"name"`
}

// ProjectRoleBindingList is the response of listing the role bindings of a Project.
type ProjectRoleBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ProjectRoleBinding `json:"items"`
}

// ProjectRoleBindingName returns the name the backend requires for the role binding of a subject:
// `cci:user:<name>` or `cci:group:<name>`, or `cci:user:<domain>:<name>` when the subject name has
// the form `<name>@<domain>`. A trailing `@` carries no domain.
func ProjectRoleBindingName(kind, subjectName string) string {
	subjectType := "user"
	if kind == ProjectRoleBindingSubjectKindGroup {
		subjectType = "group"
	}

	name, domain := subjectName, ""
	if at := strings.LastIndex(subjectName, "@"); at != -1 {
		name, domain = subjectName[:at], subjectName[at+1:]
	}
	if domain != "" {
		return fmt.Sprintf("cci:%s:%s:%s", subjectType, domain, name)
	}
	return fmt.Sprintf("cci:%s:%s", subjectType, name)
}

// NewProjectRoleBinding builds the role binding that grants role to the given subject in a Project.
// The backend can only update or delete the binding of a group whose subject name carries a
// domain separator, so a group name without one gets a trailing `@`.
func NewProjectRoleBinding(projectName, kind, subjectName, role string) ProjectRoleBinding {
	name := ProjectRoleBindingName(kind, subjectName)
	if kind == ProjectRoleBindingSubjectKindGroup && !strings.Contains(subjectName, "@") {
		subjectName += "@"
	}
	return ProjectRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       ProjectRoleBindingKind,
			APIVersion: ProjectRoleBindingAPI + "/" + ProjectRoleBindingVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: projectName,
		},
		Subjects: []ProjectRoleBindingSubject{{Kind: kind, Name: subjectName}},
		RoleRef: ProjectRoleBindingRoleRef{
			APIGroup: ProjectRoleBindingAPI,
			Kind:     ProjectRoleKind,
			Name:     role,
		},
	}
}
