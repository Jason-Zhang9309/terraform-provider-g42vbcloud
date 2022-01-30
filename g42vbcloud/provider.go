package g42vbcloud

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/helper/mutexkv"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/dcs"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/dds"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/deprecated"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/dli"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/dms"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/eip"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/evs"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/fgs"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/iam"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/rds"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/services/vpc"
)

// This is a global MutexKV for use within this plugin.
var osMutexKV = mutexkv.NewMutexKV()

// Provider returns a schema.Provider for G42VBCloud.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("G42VB_ACCESS_KEY", nil),
				Description:  descriptions["access_key"],
				RequiredWith: []string{"secret_key"},
			},

			"secret_key": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("G42VB_SECRET_KEY", nil),
				Description:  descriptions["secret_key"],
				RequiredWith: []string{"access_key"},
			},

			"security_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  descriptions["security_token"],
				RequiredWith: []string{"access_key"},
				DefaultFunc:  schema.EnvDefaultFunc("G42VB_SECURITY_TOKEN", nil),
			},

			"auth_url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc(
					"G42VB_AUTH_URL", "https://iam.ae-ad-1.vb.g42cloud.com/v3"),
				Description: descriptions["auth_url"],
			},

			"cloud": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["cloud"],
				DefaultFunc: schema.EnvDefaultFunc(
					"G42VB_CLOUD", "vb.g42cloud.com"),
			},

			"endpoints": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: descriptions["endpoints"],
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"region": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  descriptions["region"],
				DefaultFunc:  schema.EnvDefaultFunc("G42VB_REGION_NAME", nil),
				InputDefault: "ru-moscow-1",
			},

			"user_name": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("G42VB_USERNAME", ""),
				Description:  descriptions["user_name"],
				RequiredWith: []string{"password", "account_name"},
			},

			"project_name": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"G42VB_PROJECT_NAME",
				}, ""),
				Description: descriptions["project_name"],
			},

			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				DefaultFunc:  schema.EnvDefaultFunc("G42VB_PASSWORD", ""),
				Description:  descriptions["password"],
				RequiredWith: []string{"user_name", "account_name"},
			},

			"account_name": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"G42VB_ACCOUNT_NAME",
				}, ""),
				Description:  descriptions["account_name"],
				RequiredWith: []string{"password", "user_name"},
			},

			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("G42VB_INSECURE", false),
				Description: descriptions["insecure"],
			},

			"enterprise_project_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: descriptions["enterprise_project_id"],
				DefaultFunc: schema.EnvDefaultFunc("G42VB_ENTERPRISE_PROJECT_ID", ""),
			},

			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: descriptions["max_retries"],
				DefaultFunc: schema.EnvDefaultFunc("G42VB_MAX_RETRIES", 5),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"g42vbcloud_availability_zones":  huaweicloud.DataSourceAvailabilityZones(),
			"g42vbcloud_cce_cluster":         huaweicloud.DataSourceCCEClusterV3(),
			"g42vbcloud_cce_node":            huaweicloud.DataSourceCCENodeV3(),
			"g42vbcloud_cce_addon_template":  huaweicloud.DataSourceCCEAddonTemplateV3(),
			"g42vbcloud_cce_node_pool":       huaweicloud.DataSourceCCENodePoolV3(),
			"g42vbcloud_compute_flavors":     huaweicloud.DataSourceEcsFlavors(),
			"g42vbcloud_dds_flavors":         dds.DataSourceDDSFlavorV3(),
			"g42vbcloud_dcs_az":              deprecated.DataSourceDcsAZV1(),
			"g42vbcloud_dcs_flavors":         dcs.DataSourceDcsFlavorsV2(),
			"g42vbcloud_dcs_maintainwindow":  dcs.DataSourceDcsMaintainWindow(),
			"g42vbcloud_dcs_product":         deprecated.DataSourceDcsProductV1(),
			"g42vbcloud_dms_az":              deprecated.DataSourceDmsAZ(),
			"g42vbcloud_dms_product":         dms.DataSourceDmsProduct(),
			"g42vbcloud_dms_maintainwindow":  dms.DataSourceDmsMaintainWindow(),
			"g42vbcloud_identity_role":       iam.DataSourceIdentityRoleV3(),
			"g42vbcloud_images_image":        huaweicloud.DataSourceImagesImageV2(),
			"g42vbcloud_kms_key":             huaweicloud.DataSourceKmsKeyV1(),
			"g42vbcloud_kms_data_key":        huaweicloud.DataSourceKmsDataKeyV1(),
			"g42vbcloud_nat_gateway":         huaweicloud.DataSourceNatGatewayV2(),
			"g42vbcloud_networking_port":     huaweicloud.DataSourceNetworkingPortV2(),
			"g42vbcloud_networking_secgroup": huaweicloud.DataSourceNetworkingSecGroupV2(),
			"g42vbcloud_obs_bucket_object":   huaweicloud.DataSourceObsBucketObject(),
			"g42vbcloud_rds_flavors":         rds.DataSourceRdsFlavor(),
			"g42vbcloud_vpc":                 vpc.DataSourceVpcV1(),
			"g42vbcloud_vpc_bandwidth":       eip.DataSourceBandWidth(),
			"g42vbcloud_vpc_subnet":          vpc.DataSourceVpcSubnetV1(),
			"g42vbcloud_vpc_subnet_ids":      vpc.DataSourceVpcSubnetIdsV1(),
			"g42vbcloud_vpc_route":           vpc.DataSourceVpcRouteV2(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"g42vbcloud_as_configuration":          huaweicloud.ResourceASConfiguration(),
			"g42vbcloud_as_group":                  huaweicloud.ResourceASGroup(),
			"g42vbcloud_as_policy":                 huaweicloud.ResourceASPolicy(),
			"g42vbcloud_cce_cluster":               huaweicloud.ResourceCCEClusterV3(),
			"g42vbcloud_cce_node":                  huaweicloud.ResourceCCENodeV3(),
			"g42vbcloud_cce_addon":                 huaweicloud.ResourceCCEAddonV3(),
			"g42vbcloud_cce_node_pool":             huaweicloud.ResourceCCENodePool(),
			"g42vbcloud_ces_alarmrule":             huaweicloud.ResourceAlarmRule(),
			"g42vbcloud_compute_instance":          huaweicloud.ResourceComputeInstanceV2(),
			"g42vbcloud_compute_interface_attach":  huaweicloud.ResourceComputeInterfaceAttachV2(),
			"g42vbcloud_compute_keypair":           huaweicloud.ResourceComputeKeypairV2(),
			"g42vbcloud_compute_servergroup":       huaweicloud.ResourceComputeServerGroupV2(),
			"g42vbcloud_compute_eip_associate":     huaweicloud.ResourceComputeFloatingIPAssociateV2(),
			"g42vbcloud_compute_volume_attach":     huaweicloud.ResourceComputeVolumeAttachV2(),
			"g42vbcloud_dcs_instance":              dcs.ResourceDcsInstance(),
			"g42vbcloud_dds_instance":              dds.ResourceDdsInstanceV3(),
			"g42vbcloud_dli_queue":                 dli.ResourceDliQueue(),
			"g42vbcloud_dms_instance":              ResourceDmsInstancesV1(),
			"g42vbcloud_dns_recordset":             huaweicloud.ResourceDNSRecordSetV2(),
			"g42vbcloud_dns_zone":                  huaweicloud.ResourceDNSZoneV2(),
			"g42vbcloud_evs_snapshot":              huaweicloud.ResourceEvsSnapshotV2(),
			"g42vbcloud_evs_volume":                evs.ResourceEvsVolume(),
			"g42vbcloud_fgs_function":              fgs.ResourceFgsFunctionV2(),
			"g42vbcloud_identity_role_assignment":  iam.ResourceIdentityRoleAssignmentV3(),
			"g42vbcloud_identity_user":             iam.ResourceIdentityUserV3(),
			"g42vbcloud_identity_group":            iam.ResourceIdentityGroupV3(),
			"g42vbcloud_identity_group_membership": iam.ResourceIdentityGroupMembershipV3(),
			"g42vbcloud_identity_acl":              iam.ResourceIdentityACL(),
			"g42vbcloud_identity_agency":           iam.ResourceIAMAgencyV3(),
			"g42vbcloud_identity_project":          iam.ResourceIdentityProjectV3(),
			"g42vbcloud_identity_role":             iam.ResourceIdentityRole(),
			"g42vbcloud_images_image":              huaweicloud.ResourceImsImage(),
			"g42vbcloud_kms_key":                   huaweicloud.ResourceKmsKeyV1(),
			"g42vbcloud_lb_certificate":            huaweicloud.ResourceCertificateV2(),
			"g42vbcloud_lb_l7policy":               huaweicloud.ResourceL7PolicyV2(),
			"g42vbcloud_lb_l7rule":                 huaweicloud.ResourceL7RuleV2(),
			"g42vbcloud_lb_listener":               huaweicloud.ResourceListenerV2(),
			"g42vbcloud_lb_loadbalancer":           huaweicloud.ResourceLoadBalancerV2(),
			"g42vbcloud_lb_member":                 huaweicloud.ResourceMemberV2(),
			"g42vbcloud_lb_monitor":                huaweicloud.ResourceMonitorV2(),
			"g42vbcloud_lb_pool":                   huaweicloud.ResourcePoolV2(),
			"g42vbcloud_lb_whitelist":              huaweicloud.ResourceWhitelistV2(),
			"g42vbcloud_nat_dnat_rule":             huaweicloud.ResourceNatDnatRuleV2(),
			"g42vbcloud_nat_gateway":               huaweicloud.ResourceNatGatewayV2(),
			"g42vbcloud_nat_snat_rule":             huaweicloud.ResourceNatSnatRuleV2(),
			"g42vbcloud_network_acl":               huaweicloud.ResourceNetworkACL(),
			"g42vbcloud_network_acl_rule":          huaweicloud.ResourceNetworkACLRule(),
			"g42vbcloud_obs_bucket":                huaweicloud.ResourceObsBucket(),
			"g42vbcloud_obs_bucket_object":         huaweicloud.ResourceObsBucketObject(),
			"g42vbcloud_obs_bucket_policy":         huaweicloud.ResourceObsBucketPolicy(),
			"g42vbcloud_rds_instance":              ResourceRdsInstanceV3(),
			"g42vbcloud_rds_parametergroup":        huaweicloud.ResourceRdsConfigurationV3(),
			"g42vbcloud_rds_read_replica_instance": huaweicloud.ResourceRdsReadReplicaInstance(),
			"g42vbcloud_sfs_turbo":                 huaweicloud.ResourceSFSTurbo(),
			"g42vbcloud_smn_subscription":          huaweicloud.ResourceSubscription(),
			"g42vbcloud_smn_topic":                 huaweicloud.ResourceTopic(),
			"g42vbcloud_vpc":                       vpc.ResourceVirtualPrivateCloudV1(),
			"g42vbcloud_vpc_bandwidth":             eip.ResourceVpcBandWidthV2(),
			"g42vbcloud_vpc_eip":                   eip.ResourceVpcEIPV1(),
			"g42vbcloud_vpc_route":                 vpc.ResourceVPCRouteV2(),
			"g42vbcloud_vpc_peering_connection":    vpc.ResourceVpcPeeringConnectionV2(),
			"g42vbcloud_vpc_subnet":                vpc.ResourceVpcSubnetV1(),
			"g42vbcloud_networking_eip_associate":  eip.ResourceEIPAssociate(),
			"g42vbcloud_networking_secgroup":       huaweicloud.ResourceNetworkingSecGroupV2(),
			"g42vbcloud_networking_secgroup_rule":  huaweicloud.ResourceNetworkingSecGroupRuleV2(),
		},
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return configureProvider(d, terraformVersion)
	}

	return provider
}

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"auth_url": "The Identity authentication URL.",

		"cloud": "The endpoint of cloud provider, defaults to vb.g42cloud.com",

		"endpoints": "The custom endpoints used to override the default endpoint URL.",

		"region": "The G42VBCloud region to connect to.",

		"access_key": "The access key of the G42VBCloud to use.",

		"secret_key": "The secret key of the G42VBCloud to use.",

		"security_token": "The security token to authenticate with a temporary security credential.",

		"user_name": "Username to login with.",

		"project_name": "The name of the Project to login with.",

		"password": "Password to login with.",

		"account_name": "The name of the Account to login with.",

		"insecure": "Trust self-signed certificates.",
	}
}

func configureProvider(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	var project_name string

	// Use region as project_name if it's not set
	if v, ok := d.GetOk("project_name"); ok && v.(string) != "" {
		project_name = v.(string)
	} else {
		project_name = d.Get("region").(string)
	}

	config := config.Config{
		AccessKey:           d.Get("access_key").(string),
		SecretKey:           d.Get("secret_key").(string),
		SecurityToken:       d.Get("security_token").(string),
		DomainName:          d.Get("account_name").(string),
		IdentityEndpoint:    d.Get("auth_url").(string),
		Insecure:            d.Get("insecure").(bool),
		Password:            d.Get("password").(string),
		Region:              d.Get("region").(string),
		TenantName:          project_name,
		Username:            d.Get("user_name").(string),
		TerraformVersion:    terraformVersion,
		Cloud:               d.Get("cloud").(string),
		MaxRetries:          d.Get("max_retries").(int),
		EnterpriseProjectID: d.Get("enterprise_project_id").(string),
		RegionClient:        true,
		RegionProjectIDMap:  make(map[string]string),
		RPLock:              new(sync.Mutex),
	}

	if err := config.LoadAndValidate(); err != nil {
		return nil, err
	}

	if config.HwClient != nil && config.HwClient.ProjectID != "" {
		config.RegionProjectIDMap[config.Region] = config.HwClient.ProjectID
	}

	// get custom endpoints
	endpoints, err := flattenProviderEndpoints(d)
	if err != nil {
		return nil, err
	}
	config.Endpoints = endpoints

	return &config, nil
}

func flattenProviderEndpoints(d *schema.ResourceData) (map[string]string, error) {
	endpoints := d.Get("endpoints").(map[string]interface{})
	epMap := make(map[string]string)

	for key, val := range endpoints {
		endpoint := strings.TrimSpace(val.(string))
		// check empty string
		if endpoint == "" {
			return nil, fmt.Errorf("the value of customer endpoint %s must be specified", key)
		}

		// add prefix "https://" and suffix "/"
		if !strings.HasPrefix(endpoint, "http") {
			endpoint = fmt.Sprintf("https://%s", endpoint)
		}
		if !strings.HasSuffix(endpoint, "/") {
			endpoint = fmt.Sprintf("%s/", endpoint)
		}
		epMap[key] = endpoint
	}

	// unify the endpoint which has multi types
	if endpoint, ok := epMap["iam"]; ok {
		epMap["identity"] = endpoint
	}
	if endpoint, ok := epMap["ecs"]; ok {
		epMap["ecsv11"] = endpoint
		epMap["ecsv21"] = endpoint
	}
	if endpoint, ok := epMap["cce"]; ok {
		epMap["cce_addon"] = endpoint
	}
	if endpoint, ok := epMap["evs"]; ok {
		epMap["volumev2"] = endpoint
	}
	if endpoint, ok := epMap["vpc"]; ok {
		epMap["networkv2"] = endpoint
		epMap["security_group"] = endpoint
	}

	log.Printf("[DEBUG] customer endpoints: %+v", epMap)
	return epMap, nil
}
