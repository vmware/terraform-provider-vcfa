// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

// Package testutils provides shared, provider-independent helpers for VCFA acceptance
// tests. It is intentionally free of any dependency on the `vcfa` package so that it can
// be imported both by the `vcfa` package tests and by framework-based tests (which depend
// on `internal/mux`, and therefore cannot be reached from `package vcfa` without creating
// an import cycle).
package testutils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// TestConfig mirrors the JSON test configuration file used by the whole acceptance suite.
// It is defined here (rather than in vcfa/config_test.go) so that it can be shared by
// both the `vcfa` package tests (which use a type alias) and by external test packages
// such as the framework-based VKS cluster test, which cannot import `vcfa` directly.
type TestConfig struct {
	Provider struct {
		User                    string `json:"user"`
		Password                string `json:"password"`
		Token                   string `json:"token,omitempty"`
		ApiToken                string `json:"api_token,omitempty"`
		ApiTokenFile            string `json:"api_token_file,omitempty"`
		ServiceAccountTokenFile string `json:"service_account_token_file,omitempty"`

		// VCFA version and API version; allow tests to check for compatibility
		// without requiring an extra connection.
		VcfaVersion string `json:"version,omitempty"`
		ApiVersion  string `json:"apiVersion,omitempty"`

		Url                      string `json:"url"`
		SysOrg                   string `json:"sysOrg"`
		AllowInsecure            bool   `json:"allowInsecure"`
		TerraformAcceptanceTests bool   `json:"tfAcceptanceTests"`
		UseConnectionCache       bool   `json:"useConnectionCache"`
	} `json:"provider"`
	Org struct {
		Name     string `json:"name"`
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"org"`
	Cci struct {
		Region           string `json:"region"`
		Vpc              string `json:"vpc"`
		StoragePolicy    string `json:"storagePolicy"`
		SupervisorZone   string `json:"supervisorZone"`
		ContentLibrary   string `json:"contentLibrary"`
		InfraPolicyName  string `json:"infraPolicyName"`
		SharedSubnetName string `json:"sharedSubnetName"`
		VmClass1         string `json:"vmClass1"`
		VmClass2         string `json:"vmClass2"`
	} `json:"cci"`
	Vks struct {
		Project               string `json:"project"`
		Namespace             string `json:"namespace"`
		ClusterClassName      string `json:"clusterClassName"`
		ClusterClassNamespace string `json:"clusterClassNamespace"`
		KubernetesReleaseName string `json:"kubernetesReleaseName"`
		KubernetesVersion     string `json:"kubernetesVersion"`
		ServicesCidr          string `json:"servicesCidr"`
		VmClass               string `json:"vmClass"`
		StorageClass          string `json:"storageClass"`
		ControlPlaneReplicas  string `json:"controlPlaneReplicas"`
		WorkerReplicas        string `json:"workerReplicas"`
	} `json:"vks"`
	Tm struct {
		Org             string   `json:"org"`
		CreateRegion    bool     `json:"createRegion"`
		Region          string   `json:"region"`
		StorageClass    string   `json:"storageClass"`
		RegionVmClasses []string `json:"regionVmClasses"`
		ContentLibrary  string   `json:"contentLibrary"`
		Vpc             string   `json:"vpc"`

		CreateNsxManager             bool   `json:"createNsxManager"`
		NsxManagerUsername           string `json:"nsxManagerUsername"`
		NsxManagerPassword           string `json:"nsxManagerPassword"`
		NsxManagerUrl                string `json:"nsxManagerUrl"`
		NsxTier0Gateway              string `json:"nsxTier0Gateway"`
		NsxEdgeCluster               string `json:"nsxEdgeCluster"`
		NsxEdgeClusterSuffixRequired bool   `json:"nsxEdgeClusterSuffixRequired"`
		ProviderGateway              string `json:"providerGateway"`

		CreateVcenter         bool   `json:"createVcenter"`
		VcenterUsername       string `json:"vcenterUsername"`
		VcenterPassword       string `json:"vcenterPassword"`
		VcenterUrl            string `json:"vcenterUrl"`
		VcenterDatacenter     string `json:"vcenterDatacenter"`
		VcenterDatastore      string `json:"vcenterDatastore"`
		VcenterStorageProfile string `json:"vcenterStorageProfile"`
		VcenterSupervisor     string `json:"vcenterSupervisor"`
		VcenterSupervisorZone string `json:"vcenterSupervisorZone"`

		OidcServer struct {
			Url               string `json:"url,omitempty"`
			WellKnownEndpoint string `json:"wellKnownEndpoint,omitempty"`
		} `json:"oidcServer"`

		Certificates []struct {
			Path           string `json:"path,omitempty"`
			PrivateKeyPath string `json:"privateKeyPath,omitempty"`
			Password       string `json:"password,omitempty"`
		} `json:"certificates"`
		RootCertificatePath string `json:"rootCertificatePath"`
	} `json:"tm,omitempty"`
	Ldap struct {
		Host                  string `json:"host"`
		Port                  int    `json:"port"`
		IsSsl                 bool   `json:"isSsl"`
		Username              string `json:"username"`
		Password              string `json:"password"`
		BaseDistinguishedName string `json:"baseDistinguishedName"`
		Type                  string `json:"type"`
	} `json:"ldap,omitempty"`
	Logging struct {
		Enabled         bool   `json:"enabled,omitempty"`
		LogFileName     string `json:"logFileName,omitempty"`
		LogHttpRequest  bool   `json:"logHttpRequest,omitempty"`
		LogHttpResponse bool   `json:"logHttpResponse,omitempty"`
	} `json:"logging"`
	EnvVariables map[string]string `json:"envVariables,omitempty"`
}

var (
	loadOnce     sync.Once
	loadedConfig TestConfig
	loadedFile   string
)

// GetTestConfig loads (once) and returns the shared test configuration. If no configuration
// file can be located the calling test is skipped, mirroring the behaviour of the rest of
// the acceptance suite when run without a configuration file.
func GetTestConfig(t *testing.T) TestConfig {
	t.Helper()
	loadOnce.Do(func() {
		loadedFile = resolveConfigFileName()
		if loadedFile == "" {
			return
		}
		raw, err := os.ReadFile(filepath.Clean(loadedFile))
		if err != nil {
			t.Fatalf("could not read config file %s: %s", loadedFile, err)
		}
		if err := json.Unmarshal(raw, &loadedConfig); err != nil {
			t.Fatalf("could not unmarshal config file %s: %s", loadedFile, err)
		}
		applyVksDefaults(&loadedConfig)
		if TestOrgUser {
			applyOrgUserCredentials(t, &loadedConfig)
		}
	})

	if loadedFile == "" {
		t.Skipf("skipping %s: no test configuration file found (set VCFA_CONFIG)", t.Name())
	}
	return loadedConfig
}

// applyOrgUserCredentials replaces the provider credentials with the org user
// credentials from the config file. It mirrors the same substitution performed
// by vcfa/config_test.go when -vcfa-test-org-user is active.
func applyOrgUserCredentials(t *testing.T, cfg *TestConfig) {
	t.Helper()
	if cfg.Org.User == "" || cfg.Org.Password == "" {
		t.Fatalf("-vcfa-test-org-user / VCFA_TEST_ORG_USER is enabled but org user credentials (org.user / org.password) are not set in the configuration file")
	}
	cfg.Provider.User = cfg.Org.User
	cfg.Provider.Password = cfg.Org.Password
	cfg.Provider.SysOrg = cfg.Org.Name
	fmt.Println("vcfa-test-org-user enabled: using Org User credentials from configuration file")
}

// applyVksDefaults fills optional VKS fields with sensible defaults so callers do not have
// to specify them in the config file unless they want non-default values.
func applyVksDefaults(cfg *TestConfig) {
	if cfg.Vks.ClusterClassNamespace == "" {
		cfg.Vks.ClusterClassNamespace = "vmware-system-vks-public"
	}
	if cfg.Vks.ControlPlaneReplicas == "" {
		cfg.Vks.ControlPlaneReplicas = "1"
	}
	if cfg.Vks.WorkerReplicas == "" {
		cfg.Vks.WorkerReplicas = "1"
	}
}

// resolveConfigFileName locates the JSON test configuration file. Resolution order:
//  1. VCFA_CONFIG environment variable (custom location)
//  2. vcfa_test_config.json in the current working directory
//  3. <module-root>/vcfa/vcfa_test_config.json (relative to this source file)
func resolveConfigFileName() string {
	if env := os.Getenv("VCFA_CONFIG"); env != "" && fileExists(env) {
		return env
	}

	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, "vcfa_test_config.json")
		if fileExists(candidate) {
			return candidate
		}
	}

	// internal/testutils is two levels below the module root; the config lives under vcfa/
	if _, thisFile, _, ok := runtime.Caller(0); ok {
		moduleRoot := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
		candidate := filepath.Join(moduleRoot, "vcfa", "vcfa_test_config.json")
		if fileExists(candidate) {
			return candidate
		}
	}

	return ""
}

func fileExists(filename string) bool {
	info, err := os.Stat(filepath.Clean(filename))
	if err != nil {
		return false
	}
	return !info.IsDir()
}
