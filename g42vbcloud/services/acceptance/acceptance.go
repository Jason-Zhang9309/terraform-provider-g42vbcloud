package acceptance

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/g42cloud-terraform/terraform-provider-g42vbcloud/g42vbcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/logp"
)

var (
	G42VB_REGION_NAME                = os.Getenv("G42VB_REGION_NAME")
	G42VB_CUSTOM_REGION_NAME         = os.Getenv("G42VB_CUSTOM_REGION_NAME")
	G42VB_AVAILABILITY_ZONE          = os.Getenv("G42VB_AVAILABILITY_ZONE")
	G42VB_ACCESS_KEY                 = os.Getenv("G42VB_ACCESS_KEY")
	G42VB_SECRET_KEY                 = os.Getenv("G42VB_SECRET_KEY")
	G42VB_USER_ID                    = os.Getenv("G42VB_USER_ID")
	G42VB_PROJECT_ID                 = os.Getenv("G42VB_PROJECT_ID")
	G42VB_DOMAIN_ID                  = os.Getenv("G42VB_DOMAIN_ID")
	G42VB_DOMAIN_NAME                = os.Getenv("G42VB_DOMAIN_NAME")
	G42VB_ENTERPRISE_PROJECT_ID_TEST = os.Getenv("G42VB_ENTERPRISE_PROJECT_ID_TEST")

	G42VB_FLAVOR_ID        = os.Getenv("G42VB_FLAVOR_ID")
	G42VB_FLAVOR_NAME      = os.Getenv("G42VB_FLAVOR_NAME")
	G42VB_IMAGE_ID         = os.Getenv("G42VB_IMAGE_ID")
	G42VB_IMAGE_NAME       = os.Getenv("G42VB_IMAGE_NAME")
	G42VB_VPC_ID           = os.Getenv("G42VB_VPC_ID")
	G42VB_NETWORK_ID       = os.Getenv("G42VB_NETWORK_ID")
	G42VB_SUBNET_ID        = os.Getenv("G42VB_SUBNET_ID")
	G42VB_MAPREDUCE_CUSTOM = os.Getenv("G42VB_MAPREDUCE_CUSTOM")
	G42VB_ADMIN            = os.Getenv("G42VB_ADMIN")

	G42VB_DEPRECATED_ENVIRONMENT = os.Getenv("G42VB_DEPRECATED_ENVIRONMENT")

	G42VB_WAF_ENABLE_FLAG = os.Getenv("G42VB_WAF_ENABLE_FLAG")

	G42VB_DEST_REGION     = os.Getenv("G42VB_DEST_REGION")
	G42VB_DEST_PROJECT_ID = os.Getenv("G42VB_DEST_PROJECT_ID")
	G42VB_CHARGING_MODE   = os.Getenv("G42VB_CHARGING_MODE")
)

// TestAccProviders is a static map containing only the main provider instance.
//
// Deprecated: Terraform Plugin SDK version 2 uses TestCase.ProviderFactories
// but supports this value in TestCase.Providers for backwards compatibility.
// In the future Providers: TestAccProviders will be changed to
// ProviderFactories: TestAccProviderFactories
var TestAccProviders map[string]*schema.Provider

// TestAccProviderFactories is a static map containing only the main provider instance
var TestAccProviderFactories map[string]func() (*schema.Provider, error)

// TestAccProvider is the "main" provider instance
var TestAccProvider *schema.Provider

func init() {
	TestAccProvider = g42vbcloud.Provider()

	TestAccProviders = map[string]*schema.Provider{
		"g42vbcloud": TestAccProvider,
	}

	TestAccProviderFactories = map[string]func() (*schema.Provider, error){
		"g42vbcloud": func() (*schema.Provider, error) {
			return TestAccProvider, nil
		},
	}
}

// ServiceFunc the G42cloud resource query functions.
type ServiceFunc func(*config.Config, *terraform.ResourceState) (interface{}, error)

// resourceCheck resource check object, only used in the package.
type resourceCheck struct {
	resourceName    string
	resourceObject  interface{}
	getResourceFunc ServiceFunc
	resourceType    string
}

const (
	resourceTypeCode   = "resource"
	dataSourceTypeCode = "dataSource"

	checkAttrRegexpStr = `^\$\{([^\}]+)\}$`
)

/*
InitDataSourceCheck build a 'resourceCheck' object. Only used to check datasource attributes.
  Parameters:
    resourceName:    The resource name is used to check in the terraform.State.e.g. : g42vbcloud_waf_domain.domain_1.
  Return:
    *resourceCheck: resourceCheck object
*/
func InitDataSourceCheck(sourceName string) *resourceCheck {
	return &resourceCheck{
		resourceName: sourceName,
		resourceType: dataSourceTypeCode,
	}
}

/*
InitResourceCheck build a 'resourceCheck' object. The common test methods are provided in 'resourceCheck'.
  Parameters:
    resourceName:    The resource name is used to check in the terraform.State.e.g. : g42vbcloud_waf_domain.domain_1.
    resourceObject:  Resource object, used to check whether the resource exists in G42cloud.
    getResourceFunc: The function used to get the resource object.
  Return:
    *resourceCheck: resourceCheck object
*/
func InitResourceCheck(resourceName string, resourceObject interface{}, getResourceFunc ServiceFunc) *resourceCheck {
	return &resourceCheck{
		resourceName:    resourceName,
		resourceObject:  resourceObject,
		getResourceFunc: getResourceFunc,
		resourceType:    resourceTypeCode,
	}
}

func parseVariableToName(varStr string) (string, string, error) {
	var resName, keyName string
	// Check the format of the variable.
	match, _ := regexp.MatchString(checkAttrRegexpStr, varStr)
	if !match {
		return resName, keyName, fmtp.Errorf("The type of 'variable' is error, "+
			"expected ${resourceType.name.field} got %s", varStr)
	}

	reg, err := regexp.Compile(checkAttrRegexpStr)
	if err != nil {
		return resName, keyName, fmtp.Errorf("The acceptance function is wrong.")
	}
	mArr := reg.FindStringSubmatch(varStr)
	if len(mArr) != 2 {
		return resName, keyName, fmtp.Errorf("The type of 'variable' is error, "+
			"expected ${resourceType.name.field} got %s", varStr)
	}

	// Get resName and keyName from variable.
	strs := strings.Split(mArr[1], ".")
	for i, s := range strs {
		if strings.Contains(s, "g42vbcloud_") {
			resName = strings.Join(strs[0:i+2], ".")
			keyName = strings.Join(strs[i+2:], ".")
			break
		}
	}
	return resName, keyName, nil
}

/*
TestCheckResourceAttrWithVariable validates the variable in state for the given name/key combination.
  Parameters:
    resourceName: The resource name is used to check in the terraform.State.
    key:          The field name of the resource.
    variable:     The variable name of the value to be checked.

    variable such like ${g42vbcloud_waf_certificate.certificate_1.id}
    or ${data.g42vbcloud_waf_policies.policies_2.policies.0.id}
*/
func TestCheckResourceAttrWithVariable(resourceName, key, varStr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resName, keyName, err := parseVariableToName(varStr)
		if err != nil {
			return err
		}

		if strings.EqualFold(resourceName, resName) {
			return fmtp.Errorf("Meaningless verification. " +
				"The referenced resource cannot be the current resource.")
		}

		// Get the value based on resName and keyName from the state.
		rs, ok := s.RootModule().Resources[resName]
		if !ok {
			return fmtp.Errorf("Can't find %s in state : %s.", resName, ok)
		}
		value := rs.Primary.Attributes[keyName]

		return resource.TestCheckResourceAttr(resourceName, key, value)(s)
	}
}

// CheckResourceDestroy check whether resources destroied in G42cloud.
func (rc *resourceCheck) CheckResourceDestroy() resource.TestCheckFunc {
	if strings.Compare(rc.resourceType, dataSourceTypeCode) == 0 {
		fmtp.Errorf("Error, you built a resourceCheck with 'InitDataSourceCheck', " +
			"it cannot run CheckResourceDestroy().")
		return nil
	}
	return func(s *terraform.State) error {
		strs := strings.Split(rc.resourceName, ".")
		var resourceType string
		for _, str := range strs {
			if strings.Contains(str, "g42vbcloud_") {
				resourceType = strings.Trim(str, " ")
				break
			}
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			conf := TestAccProvider.Meta().(*config.Config)
			if rc.getResourceFunc != nil {
				if _, err := rc.getResourceFunc(conf, rs); err == nil {
					return fmtp.Errorf("failed to destroy resource. The resource of %s : %s still exists.",
						resourceType, rs.Primary.ID)
				}
			} else {
				return fmtp.Errorf("The 'getResourceFunc' is nil, please set it during initialization.")
			}
		}
		return nil
	}
}

// CheckResourceExists check whether resources exist in G42cloud.
func (rc *resourceCheck) CheckResourceExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rc.resourceName]
		if !ok {
			return fmtp.Errorf("Can not found the resource or data source in state: %s", rc.resourceName)
		}
		if rs.Primary.ID == "" {
			return fmtp.Errorf("No id set for the resource or data source: %s", rc.resourceName)
		}
		if strings.EqualFold(rc.resourceType, dataSourceTypeCode) {
			return nil
		}

		if rc.getResourceFunc != nil {
			conf := TestAccProvider.Meta().(*config.Config)
			r, err := rc.getResourceFunc(conf, rs)
			if err != nil {
				return fmtp.Errorf("checking resource %s %s exists error: %s ",
					rc.resourceName, rs.Primary.ID, err)
			}
			if rc.resourceObject != nil {
				rc.resourceObject = r
			} else {
				logp.Printf("[WARN] The 'resourceObject' is nil, please set it during initialization.")
			}
		} else {
			return fmtp.Errorf("The 'getResourceFunc' is nil, please set it.")
		}

		return nil
	}
}

func preCheckRequiredEnvVars(t *testing.T) {
	if G42VB_REGION_NAME == "" {
		t.Fatal("G42VB_REGION_NAME must be set for acceptance tests")
	}
}

//lintignore:AT003
func TestAccPreCheck(t *testing.T) {
	// Do not run the test if this is a deprecated testing environment.
	if G42VB_DEPRECATED_ENVIRONMENT != "" {
		t.Skip("This environment only runs deprecated tests")
	}

	preCheckRequiredEnvVars(t)
}

//lintignore:AT003
func TestAccPrecheckCustomRegion(t *testing.T) {
	if G42VB_CUSTOM_REGION_NAME == "" {
		t.Skip("G42VB_CUSTOM_REGION_NAME must be set for acceptance tests")
	}
}

//lintignore:AT003
func TestAccPreCheckDeprecated(t *testing.T) {
	if G42VB_DEPRECATED_ENVIRONMENT == "" {
		t.Skip("This environment does not support deprecated tests")
	}

	preCheckRequiredEnvVars(t)
}

//lintignore:AT003
func TestAccPreCheckEpsID(t *testing.T) {
	// use G42VB_ENTERPRISE_PROJECT_ID_TEST instead of G42VB_ENTERPRISE_PROJECT_ID to avoid enabling EPS globally
	if G42VB_ENTERPRISE_PROJECT_ID_TEST == "" {
		t.Skip("This environment does not support Enterprise Project ID tests")
	}
}

//lintignore:AT003
func TestAccPreCheckBms(t *testing.T) {
	if G42VB_USER_ID == "" {
		t.Skip("G42VB_USER_ID must be set for BMS acceptance tests")
	}
}

//lintignore:AT003
func TestAccPreCheckMrsCustom(t *testing.T) {
	if G42VB_MAPREDUCE_CUSTOM == "" {
		t.Skip("G42VB_MAPREDUCE_CUSTOM must be set for acceptance tests:custom type cluster of map reduce")
	}
}

func RandomAccResourceName() string {
	return fmt.Sprintf("tf_acc_test_%s", acctest.RandString(5))
}

func RandomAccResourceNameWithDash() string {
	return fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
}

func RandomCidr() string {
	return fmt.Sprintf("172.16.%d.0/24", acctest.RandIntRange(0, 255))
}

func RandomCidrAndGatewayIp() (string, string) {
	seed := acctest.RandIntRange(0, 255)
	return fmt.Sprintf("172.16.%d.0/24", seed), fmt.Sprintf("172.16.%d.1", seed)
}

//lintignore:AT003
func TestAccPrecheckWafInstance(t *testing.T) {
	if G42VB_WAF_ENABLE_FLAG == "" {
		t.Skip("Jump the WAF acceptance tests.")
	}
}

//lintignore:AT003
func TestAccPreCheckAdminOnly(t *testing.T) {
	if G42VB_ADMIN == "" {
		t.Skip("Skipping test because it requires the admin privileges")
	}
}

//lintignore:AT003
func TestAccPreCheckReplication(t *testing.T) {
	if G42VB_DEST_REGION == "" || G42VB_DEST_PROJECT_ID == "" {
		t.Skip("Jump the replication policy acceptance tests.")
	}
}

//lintignore:AT003
func TestAccPreCheckProject(t *testing.T) {
	if G42VB_ENTERPRISE_PROJECT_ID_TEST != "" {
		t.Skip("This environment does not support project tests")
	}
}

//lintignore:AT003
func TestAccPreCheckOBS(t *testing.T) {
	if G42VB_ACCESS_KEY == "" || G42VB_SECRET_KEY == "" {
		t.Skip("G42VB_ACCESS_KEY and G42VB_SECRET_KEY must be set for OBS acceptance tests")
	}
}

//lintignore:AT003
func TestAccPreCheckChargingMode(t *testing.T) {
	if G42VB_CHARGING_MODE != "prePaid" {
		t.Skip("This environment does not support prepaid tests")
	}
}
