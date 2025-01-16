package vcfa

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

func init() {
	separator := os.Getenv("VCFA_IMPORT_SEPARATOR")
	if separator != "" {
		ImportSeparator = separator
	}
}

type Config struct {
	User                    string
	Password                string
	Token                   string // Token used instead of user and password
	ApiToken                string // User generated token used instead of user and password
	ApiTokenFile            string // File containing a user generated API token
	AllowApiTokenFile       bool   // Setting to suppress API Token File security warnings
	ServiceAccountTokenFile string // File containing the Service Account API token
	AllowSATokenFile        bool   // Setting to suppress Service Account Token File security warnings
	SysOrg                  string // Org used for authentication
	Org                     string // Default Org used for API operations
	Vdc                     string // Default (optional) VDC for API operations
	Href                    string
	InsecureFlag            bool
}

type VCDClient struct {
	*govcd.VCDClient
	SysOrg       string
	Org          string // name of default Org
	Vdc          string // name of default VDC
	InsecureFlag bool
}

// StringMap type is used to simplify reading resource definitions
type StringMap map[string]interface{}

// Cache values for VCFA connection.
// When the Client() function is called with the same parameters, it will return
// a cached value instead of connecting again.
// This makes the Client() function both deterministic and fast.
//
// WARNING: Cached clients need to be evicted by calling cacheStorage.reset() after the rights of the associated
// logged user change. Otherwise, retrieving or manipulating objects that require the new rights could return 403
// forbidden errors. For example, adding read rights of a RDE Type to a specific user requires a cacheStorage.reset()
// afterwards, to force a re-authentication. If this is not done, the cached client won't be able to read this RDE Type.
type cachedConnection struct {
	initTime   time.Time
	connection *VCDClient
}

type cacheStorage struct {
	// conMap holds cached VCFA authenticated connection
	conMap map[string]cachedConnection
	// cacheClientServedCount records how many times we have cached a connection
	cacheClientServedCount int
	sync.Mutex
}

var (
	// Enables the caching of authenticated connections
	enableConnectionCache = os.Getenv("VCFA_CACHE") != ""

	// Cached VCFA authenticated connection
	cachedVCDClients = &cacheStorage{conMap: make(map[string]cachedConnection)}

	// Invalidates the cache after a given time (connection tokens usually expire after 20 to 30 minutes)
	maxConnectionValidity = 20 * time.Minute

	enableDebug = os.Getenv("GOVCD_DEBUG") != ""
	enableTrace = os.Getenv("GOVCD_TRACE") != ""

	// ImportSeparator is the separation string used for import operations
	// Can be changed using either "import_separator" property in Provider
	// or environment variable "VCFA_IMPORT_SEPARATOR"
	ImportSeparator = "."
)

// Displays conditional messages
func debugPrintf(format string, args ...interface{}) {
	// When GOVCD_TRACE is enabled, we also display the function that generated the message
	if enableTrace {
		format = fmt.Sprintf("[%s] %s", filepath.Base(callFuncName()), format)
	}
	// The formatted message passed to this function is displayed only when GOVCD_DEBUG is enabled.
	if enableDebug {
		fmt.Printf(format, args...)
	}
}

// TODO Look into refactoring this into a method of *Config
func ProviderAuthenticate(client *govcd.VCDClient, user, password, token, org, apiToken, apiTokenFile, saTokenFile string) error {
	var err error
	if saTokenFile != "" {
		return client.SetServiceAccountApiToken(org, saTokenFile)
	}
	if apiTokenFile != "" {
		_, err := client.SetApiTokenFromFile(org, apiTokenFile)
		if err != nil {
			return err
		}
		return nil
	}
	if apiToken != "" {
		return client.SetToken(org, govcd.ApiTokenHeader, apiToken)
	}
	if token != "" {
		if len(token) > 32 {
			err = client.SetToken(org, govcd.BearerTokenHeader, token)
		} else {
			err = client.SetToken(org, govcd.AuthorizationHeader, token)
		}
		if err != nil {
			return fmt.Errorf("error during token-based authentication: %s", err)
		}
		return nil
	}

	return client.Authenticate(user, password, org)
}

func (c *Config) Client() (*VCDClient, error) {
	rawData := c.User + "#" +
		c.Password + "#" +
		c.Token + "#" +
		c.ApiToken + "#" +
		c.ApiTokenFile + "#" +
		c.ServiceAccountTokenFile + "#" +
		c.SysOrg + "#" +
		c.Vdc + "#" +
		c.Href
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(rawData)))

	// The cached connection is served only if the variable VCFA_CACHE is set
	cachedVCDClients.Lock()
	client, ok := cachedVCDClients.conMap[checksum]
	cachedVCDClients.Unlock()
	if ok && enableConnectionCache {
		cachedVCDClients.Lock()
		cachedVCDClients.cacheClientServedCount += 1
		cachedVCDClients.Unlock()
		// debugPrintf("[%s] cached connection served %d times (size:%d)\n",
		elapsed := time.Since(client.initTime)
		if elapsed > maxConnectionValidity {
			debugPrintf("cached connection invalidated after %2.0f minutes \n", maxConnectionValidity.Minutes())
			cachedVCDClients.Lock()
			delete(cachedVCDClients.conMap, checksum)
			cachedVCDClients.Unlock()
		} else {
			return client.connection, nil
		}
	}

	authUrl, err := url.ParseRequestURI(c.Href)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while retrieving URL: %s", err)
	}

	userAgent := buildUserAgent(BuildVersion, c.SysOrg)

	vcdClient := &VCDClient{
		VCDClient: govcd.NewVCDClient(*authUrl, c.InsecureFlag,
			govcd.WithHttpUserAgent(userAgent),
		),
		SysOrg:       c.SysOrg,
		Org:          c.Org,
		Vdc:          c.Vdc,
		InsecureFlag: c.InsecureFlag}

	err = ProviderAuthenticate(vcdClient.VCDClient, c.User, c.Password, c.Token, c.SysOrg, c.ApiToken, c.ApiTokenFile, c.ServiceAccountTokenFile)
	if err != nil {
		return nil, fmt.Errorf("something went wrong during authentication: %s", err)
	}
	cachedVCDClients.Lock()
	cachedVCDClients.conMap[checksum] = cachedConnection{initTime: time.Now(), connection: vcdClient}
	cachedVCDClients.Unlock()

	return vcdClient, nil
}

// callFuncName returns the name of the function that called the current function. It is used for
// tracing
func callFuncName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)
	if n > 0 {
		fun := runtime.FuncForPC(fpcs[0] - 1)
		if fun != nil {
			return fun.Name()
		}
	}
	return ""
}

// buildUserAgent helps to construct HTTP User-Agent header
func buildUserAgent(version, sysOrg string) string {
	userAgent := fmt.Sprintf("terraform-provider-vcfa/%s (%s/%s; isProvider:%t)",
		version, runtime.GOOS, runtime.GOARCH, strings.ToLower(sysOrg) == "system")

	return userAgent
}

// dSet sets the value of a schema property, discarding the error
// Use only for scalar values (strings, booleans, and numbers)
func dSet(d *schema.ResourceData, key string, value interface{}) {
	if value != nil && !isScalar(value) {
		msg1 := "*** ERROR: only scalar values should be used for dSet()"
		msg2 := fmt.Sprintf("*** detected '%s' for key '%s' (called from %s)",
			reflect.TypeOf(value).Kind(), key, callFuncName())
		starLine := strings.Repeat("*", len(msg2))
		// This panic should never reach the final user.
		// Its purpose is to alert the developer that there was an improper use of `dSet`
		panic(fmt.Sprintf("\n%s\n%s\n%s\n%s\n", starLine, msg1, msg2, starLine))
	}
	err := d.Set(key, value)
	if err != nil {
		panic(fmt.Sprintf("error in %s - key '%s': %s ", callFuncName(), key, err))
	}
}

// isScalar returns true if its argument is not a composite object
// we want strings, numbers, booleans
func isScalar(t interface{}) bool {
	if t == nil {
		return true
	}
	typeOf := reflect.TypeOf(t)
	switch typeOf.Kind().String() {
	case "struct", "map", "array", "slice":
		return false
	}

	return true
}
