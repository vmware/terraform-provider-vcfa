//go:build api || functional || tm || ALL

package vcfa

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	testingTags["api"] = "provider_test.go"
}

// testAccProvider is a global provider used in tests
var testAccProvider *schema.Provider

// testAccProviders used in field ProviderFactories required for test runs in SDK 2.x
var testAccProviders map[string]func() (*schema.Provider, error)

func TestProvider(t *testing.T) {
	// Do not add pre and post checks
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	// Do not add pre and post checks
	var _ *schema.Provider = Provider()
}

// When this function is called, the initialization in config_test.go has already happened.
// Therefore, we can safely require that testConfig fields used in test params have been filled.
// Note: This function call moved from resource.Test.PreCheck to before templateFill function call to avoid generation
// of binary test in case values are missing
func testParamsNotEmpty(t *testing.T, params StringMap) {
	for key, value := range params {
		if value == "" {
			t.Skipf("[%s] %s must be set for acceptance tests", t.Name(), key)
		}
	}
}

// createTemporaryVCFAConnection is meant to create a VCDClient to check environment before executing specific acceptance
// tests and before VCDClient is accessible.
func createTemporaryVCFAConnection(acceptNil bool) *VCDClient {
	config := Config{
		User:            testConfig.Provider.User,
		Password:        testConfig.Provider.Password,
		Token:           testConfig.Provider.Token,
		ApiToken:        testConfig.Provider.ApiToken,
		UseSamlAdfs:     testConfig.Provider.UseSamlAdfs,
		CustomAdfsRptId: testConfig.Provider.CustomAdfsRptId,
		SysOrg:          testConfig.Provider.SysOrg,
		Org:             testConfig.VCD.Org,
		Vdc:             testConfig.Nsxt.Vdc,
		Href:            testConfig.Provider.Url,
		InsecureFlag:    testConfig.Provider.AllowInsecure,
		MaxRetryTimeout: testConfig.Provider.MaxRetryTimeout,
	}
	conn, err := config.Client()
	if err != nil {
		if acceptNil {
			return nil
		}
		panic("unable to initialize VCFA connection :" + err.Error())
	}
	return conn
}

// createSystemTemporaryVCFAConnection is like createTemporaryVCFAConnection, but it will ignore all conditional
// configurations like `VCFA_TEST_ORG_USER=1` and will still return a System client instead of user one. This allows to
// perform System actions (entities which require System rights - Org, Vdc, etc...)
func createSystemTemporaryVCFAConnection() *VCDClient {
	var configStruct TestConfig
	configFileName := getConfigFileName()

	// Looks if the configuration file exists before attempting to read it
	if configFileName == "" {
		panic(fmt.Errorf("configuration file %s not found", configFileName))
	}
	jsonFile, err := os.ReadFile(filepath.Clean(configFileName))
	if err != nil {
		panic(fmt.Errorf("could not read config file %s: %v", configFileName, err))
	}
	err = json.Unmarshal(jsonFile, &configStruct)
	if err != nil {
		panic(fmt.Errorf("could not unmarshal json file: %v", err))
	}

	config := Config{
		User:            configStruct.Provider.User,
		Password:        configStruct.Provider.Password,
		Token:           configStruct.Provider.Token,
		UseSamlAdfs:     configStruct.Provider.UseSamlAdfs,
		CustomAdfsRptId: configStruct.Provider.CustomAdfsRptId,
		SysOrg:          configStruct.Provider.SysOrg,
		Org:             configStruct.VCD.Org,
		Vdc:             configStruct.Nsxt.Vdc,
		Href:            configStruct.Provider.Url,
		InsecureFlag:    configStruct.Provider.AllowInsecure,
		MaxRetryTimeout: configStruct.Provider.MaxRetryTimeout,
	}
	conn, err := config.Client()
	if err != nil {
		panic("unable to initialize VCFA connection :" + err.Error())
	}
	return conn
}

// minIfLess returns:
// `min` if `value` is less than min
// `value` if `value` > `min`
func minIfLess(min, value int) int {
	if value < min {
		return min
	}

	return value
}

// TestAccClientUserAgent ensures that client initialization config.Client() used by provider initializes
// go-vcloud-director client by having User-Agent set
func TestAccClientUserAgent(t *testing.T) {
	// Do not add pre and post checks
	// Exit the test early
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	clientConfig := Config{
		User:            testConfig.Provider.User,
		Password:        testConfig.Provider.Password,
		Token:           testConfig.Provider.Token,
		SysOrg:          testConfig.Provider.SysOrg,
		Org:             testConfig.VCD.Org,
		Vdc:             testConfig.Nsxt.Vdc,
		Href:            testConfig.Provider.Url,
		MaxRetryTimeout: testConfig.Provider.MaxRetryTimeout,
		InsecureFlag:    testConfig.Provider.AllowInsecure,
	}

	vcdClient, err := clientConfig.Client()
	if err != nil {
		t.Fatal("error initializing go-vcloud-director client: " + err.Error())
	}

	expectedHeaderPrefix := "terraform-provider-vcfa/"
	if !strings.HasPrefix(vcdClient.VCDClient.Client.UserAgent, expectedHeaderPrefix) {
		t.Fatalf("Expected User-Agent header in go-vcloud-director to be '%s', got '%s'",
			expectedHeaderPrefix, vcdClient.VCDClient.Client.UserAgent)
	}
}

// createOrgVCFAConnection creates a connection to the VCFA using an Org user
// The credentials are the ones set in the configuration file for the org user.
// Passing an empty suffix will connect to the first Org (example: "testorg")
// Passing a suffix "-1" will connect to the second org (example: "testorg-1")
func createOrgVCFAConnection(orgSuffix string) *VCDClient {
	config := Config{
		User:            testConfig.TestEnvBuild.OrgUser,
		Password:        testConfig.TestEnvBuild.OrgUserPassword,
		Token:           "",
		ApiToken:        "",
		UseSamlAdfs:     false,
		CustomAdfsRptId: "",
		SysOrg:          testConfig.VCD.Org + orgSuffix,
		Org:             testConfig.VCD.Org + orgSuffix,
		Vdc:             "",
		Href:            testConfig.Provider.Url,
		InsecureFlag:    testConfig.Provider.AllowInsecure,
		MaxRetryTimeout: testConfig.Provider.MaxRetryTimeout,
	}
	conn, err := config.Client()
	if err != nil {
		panic("unable to initialize VCFA connection :" + err.Error())
	}
	return conn
}

// testOrgProvider creates a new provider with Org credentials
// See createOrgVCFAConnection to see how to use orgSuffix
func testOrgProvider(orgSuffix string) *schema.Provider {
	newProvider := Provider()

	newProvider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return createOrgVCFAConnection(orgSuffix), nil
	}
	return newProvider
}

// buildMultipleProviders builds a provider factory with a system administrator and
// two Org users, taking the credentials from the configuration file
func buildMultipleProviders() map[string]func() (*schema.Provider, error) {
	providers := map[string]func() (*schema.Provider, error){
		providerVcfaSystem: func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
		providerVcfaOrg1: func() (*schema.Provider, error) {
			return testOrgProvider(""), nil
		},
		providerVcfaOrg2: func() (*schema.Provider, error) {
			return testOrgProvider("-1"), nil
		},
	}
	return providers
}

// ----------------------
func createSysVCFAConnection(vcfaUrl, user, password, org string) *VCDClient {
	config := Config{
		User:            user,
		Password:        password,
		Token:           "",
		ApiToken:        "",
		UseSamlAdfs:     false,
		CustomAdfsRptId: "",
		SysOrg:          org,
		Org:             org,
		Vdc:             "",
		Href:            vcfaUrl,
		InsecureFlag:    testConfig.Provider.AllowInsecure,
		MaxRetryTimeout: testConfig.Provider.MaxRetryTimeout,
	}
	conn, err := config.Client()
	if err != nil {
		panic("unable to initialize VCFA connection :" + err.Error())
	}
	return conn
}

// buildMultipleSysProviders builds a provider factory with two system administrators from two VCFAs
func buildMultipleSysProviders(vcfaUrl, user, password, org string) map[string]func() (*schema.Provider, error) {
	newProvider := Provider()

	newProvider.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return createSysVCFAConnection(vcfaUrl, user, password, org), nil
	}
	providers := map[string]func() (*schema.Provider, error){
		providerVcfaSystem: func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
		providerVcfaSystem2: func() (*schema.Provider, error) {
			return newProvider, nil
		},
	}
	return providers
}
