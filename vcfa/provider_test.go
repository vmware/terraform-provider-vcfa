//go:build api || functional || tm || ALL

package vcfa

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
		User:         configStruct.Provider.User,
		Password:     configStruct.Provider.Password,
		Token:        configStruct.Provider.Token,
		SysOrg:       configStruct.Provider.SysOrg,
		Org:          configStruct.Provider.SysOrg,
		Href:         configStruct.Provider.Url,
		InsecureFlag: configStruct.Provider.AllowInsecure,
	}
	conn, err := config.Client()
	if err != nil {
		panic("unable to initialize VCFA connection :" + err.Error())
	}
	return conn
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
		User:         testConfig.Provider.User,
		Password:     testConfig.Provider.Password,
		Token:        testConfig.Provider.Token,
		SysOrg:       testConfig.Provider.SysOrg,
		Org:          testConfig.Provider.SysOrg,
		Href:         testConfig.Provider.Url,
		InsecureFlag: testConfig.Provider.AllowInsecure,
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
