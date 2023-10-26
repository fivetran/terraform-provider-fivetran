package fivetran

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/destinations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDestinationCreate,
		ReadContext:          resourceDestinationRead,
		UpdateWithoutTimeout: resourceDestinationUpdate,
		DeleteContext:        resourceDestinationDelete,
		Importer:             &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: 			  getDestinationSchema(false),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func getDestinationSchema(datasource bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    !datasource,
			Required:    datasource,
			Description: "The unique identifier for the destination within the Fivetran system.",
		},
		"group_id": {
			Type:        schema.TypeString,
			Required:    !datasource,
			ForceNew:    !datasource,
			Computed:    datasource,
			Description: "The unique identifier for the Group within the Fivetran system.",
		},
		"service": {
			Type:        schema.TypeString,
			Required:    !datasource,
			ForceNew:    !datasource,
			Computed:    datasource,
			Description: "The destination type name within the Fivetran system.",
		},
		"region": {
			Type:        schema.TypeString,
			Required:    !datasource,
			Computed:    datasource,
			Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
		},
		"time_zone_offset": {
			Type:        schema.TypeString,
			Required:    !datasource,
			Computed:    datasource,
			Description: "Determines the time zone for the Fivetran sync schedule.",
		},
		"config": getDestinationSchemaConfig(datasource),
		"trust_certificates": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Specifies whether we should trust the certificate automatically. The default value is FALSE. If a certificate is not trusted automatically, it has to be approved with [Certificates Management API Approve a destination certificate](https://fivetran.com/docs/rest-api/certificates#approveadestinationcertificate).",
		},
		"trust_fingerprints": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Specifies whether we should trust the SSH fingerprint automatically. The default value is FALSE. If a fingerprint is not trusted automatically, it has to be approved with [Certificates Management API Approve a destination fingerprint](https://fivetran.com/docs/rest-api/certificates#approveadestinationfingerprint).",
		},
		"run_setup_tests": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     datasource,
			Description: "Specifies whether the setup tests should be run automatically. The default value is TRUE.",
		},
		"setup_status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Destination setup status",
		},
		"last_updated": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "",
		},
	}
}

func getDestinationSchemaConfig(datasource bool) *schema.Schema {
	maxItems := 1
	if datasource {
		maxItems = 0
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Required: !datasource,
		Optional: datasource,
		Computed: datasource,
		MaxItems: maxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"host": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Server name",
				},
				"port": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "Server port number",
				},
				"database": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Database name",
				},
				"auth": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The connector authorization settings. Check possible config formats in [create method](/openapi/reference/v1/operation/create_connector/)",
				},
				"user": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Database user name",
				},
				"password": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "Database user password",
				},
				"connection_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Connection method. Default value: `Directly`.",
				},
				"tunnel_host": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "SSH server name. Must be populated if `connection_type` is set to `SshTunnel`.",
				},
				"tunnel_port": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "SSH server port name. Must be populated if `connection_type` is set to `SshTunnel`.",
				},
				"tunnel_user": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "SSH user name. Must be populated if `connection_type` is set to `SshTunnel`.",
				},
				"project_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "BigQuery project ID",
				},
				"data_set_location": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Data location. Datasets will reside in this location.",
				},
				"bucket": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Customer bucket. If specified, your GCS bucket will be used to process the data instead of a Fivetran-managed bucket. The bucket must be present in the same location as the dataset location.",
				},
				"server_host_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Server name",
				},
				"http_path": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "HTTP path",
				},
				"personal_access_token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "Personal access token",
				},
				"create_external_tables": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Whether to create external tables",
				},
				"external_location": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "External location to store Delta tables. Default value: `\"\"`  (null). By default, the external tables will reside in the `/{schema}/{table}` path, and if you specify an external location in the `{externalLocation}/{schema}/{table}` path.",
				},
				"auth_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Authentication type. Default value: `PASSWORD`.",
				},
				"role_arn": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "Role ARN with Redshift permissions. Required if authentication type is `IAM`.",
				},
				"secret_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "Private key of the customer service account. If specified, your service account will be used to process the data instead of the Fivetran-managed service account.",
				},
				"private_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "Private access key.  The field should be specified if authentication type is `KEY_PAIR`.",
				},
				"public_key": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Public key to grant Fivetran SSH access to git repository.",
				},
				"cluster_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Cluster ID. Must be populated if `connection_type` is set to `SshTunnel` and `auth_type` is set to `IAM`.",
				},
				"cluster_region": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Cluster region. Must be populated if `connection_type` is set to `SshTunnel` and `auth_type` is set to `IAM`.",
				},
				"role": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The group role that you would like to assign this new user to. Supported group roles: ‘Destination Administrator‘, ‘Destination Reviewer‘, ‘Destination Analyst‘, ‘Connector Creator‘, or a custom destination role",
				},
				"is_private_key_encrypted": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "Indicates that a private key is encrypted. The default value: `false`. The field can be specified if authentication type is `KEY_PAIR`.",
				},
				"passphrase": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "In case private key is encrypted, you are required to enter passphrase that was used to encrypt the private key. The field can be specified if authentication type is `KEY_PAIR`.",
				},
				"catalog": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Catalog name",
				},
				"fivetran_role_arn": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "ARN of the role which you created with different required policy mentioned in our setup guide",
				},
				"prefix_path": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Prefix path of the bucket for which you have configured access policy. It is not required if access has been granted to entire Bucket in the access policy",
				},
				"region": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Region of your AWS S3 bucket",
				},
			},
		},
	}
}

func resourceDestinationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDestinationCreate()

	svc.GroupID(d.Get("group_id").(string))
	svc.Service(d.Get("service").(string))
	svc.Region(d.Get("region").(string))
	svc.TimeZoneOffset(d.Get("time_zone_offset").(string))
	if v, ok := resourceDestinationCreateConfig(d.Get("config").([]interface{})); ok {
		svc.Config(v)
	}
	if v, ok := d.GetOk("trust_certificates"); ok {
		svc.TrustCertificates(v.(bool))
	}
	if v, ok := d.GetOk("trust_fingerprints"); ok {
		svc.TrustFingerprints(v.(bool))
	}
	if v, ok := d.GetOk("run_setup_tests"); ok {
		svc.RunSetupTests(v.(bool))
	}

	ctx, cancel := setContextTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceDestinationRead(ctx, d, m)

	return diags
}

func resourceDestinationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDestinationDetails()

	resp, err := svc.DestinationID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		// If the resource does not exist (404), inform Terraform. We want to immediately
		// return here to prevent further processing.
		if resp.Code == "404" {
			d.SetId("")
			return nil
		}
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.ID
	msi["group_id"] = resp.Data.GroupID
	msi["service"] = resp.Data.Service
	msi["region"] = resp.Data.Region
	msi["time_zone_offset"] = resp.Data.TimeZoneOffset

	config, err := resourceDestinationReadConfig(&resp, d.Get("config").([]interface{}))
	if err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}
	msi["config"] = config
	msi["setup_status"] = resp.Data.SetupStatus
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func resourceDestinationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDestinationModify()

	svc.DestinationID(d.Get("id").(string))

	hasChanges := false

	if d.HasChange("region") {
		svc.Region(d.Get("region").(string))
		hasChanges = true
	}
	if d.HasChange("time_zone_offset") {
		svc.TimeZoneOffset(d.Get("time_zone_offset").(string))
		hasChanges = true
	}
	if d.HasChange("config") {
		_, n := d.GetChange("config")
		// resourceDestinationCreateConfig is used here because
		// the whole "config" block must be sent to the REST API.
		if v, ok := resourceDestinationCreateConfig(n.([]interface{})); ok {
			svc.Config(v)
			hasChanges = true
			// only sets change if func resourceDestinationCreateConfig returns ok
		}
	}
	ctx, cancel := setContextTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	if hasChanges {
		if v, ok := d.GetOk("run_setup_tests"); ok {
			svc.RunSetupTests(v.(bool))
		}

		resp, err := svc.Do(ctx)
		if err != nil {
			// resourceDestinationRead here makes sure the state is updated after a NewDestinationModify error.
			diags = resourceDestinationRead(ctx, d, m)
			return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
		}

		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	} else {
		// if only "run_setup_tests" updated to true - setup tests should be performed without update request
		if v, ok := d.GetOk("run_setup_tests"); ok && v.(bool) && d.HasChange("run_setup_tests") {
			testsSvc := client.NewDestinationSetupTests().DestinationID(d.Get("id").(string))
			if v, ok := d.GetOk("trust_certificates"); ok {
				testsSvc.TrustCertificates(v.(bool))
			}
			if v, ok := d.GetOk("trust_fingerprints"); ok {
				testsSvc.TrustFingerprints(v.(bool))
			}
			resp, err := testsSvc.Do(ctx)
			if err != nil {
				return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
			}
		}
	}

	return resourceDestinationRead(ctx, d, m)
}

func resourceDestinationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDestinationDelete()

	resp, err := svc.DestinationID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}

// resourceDestinationReadConfig receives a *fivetran.DestinationDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" set.
func resourceDestinationReadConfig(resp *destinations.DestinationDetailsResponse, currentConfig []interface{}) ([]interface{}, error) {
	var config []interface{}

	c := make(map[string]interface{})
	c["host"] = resp.Data.Config.Host
	if resp.Data.Config.Port != "" {
		port, err := strconv.Atoi(resp.Data.Config.Port)
		if err != nil {
			return config, err
		}
		c["port"] = port
	}
	c["database"] = resp.Data.Config.Database
	c["auth"] = resp.Data.Config.Auth
	c["user"] = resp.Data.Config.User

	if len(currentConfig) > 0 {
		// The REST API sends the password field masked. We use the state stored password here if possible.
		currentConfigMap := currentConfig[0].(map[string]interface{})
		c["password"] = currentConfigMap["password"].(string)
		c["private_key"] = currentConfigMap["private_key"].(string)
		c["secret_key"] = currentConfigMap["secret_key"].(string)
		c["personal_access_token"] = currentConfigMap["personal_access_token"].(string)
		c["role_arn"] = currentConfigMap["role_arn"].(string)
		c["passphrase"] = currentConfigMap["passphrase"].(string)

		if _, ok := currentConfigMap["is_private_key_encrypted"]; ok {
			// if `is_private_key_encrypted` is configured locally we should read upstream value
			c["is_private_key_encrypted"] = resp.Data.Config.IsPrivateKeyEncrypted
		}
	} else {
		c["password"] = resp.Data.Config.Password
		c["personal_access_token"] = resp.Data.Config.PersonalAccessToken
		c["role_arn"] = resp.Data.Config.RoleArn
		c["secret_key"] = resp.Data.Config.SecretKey
		c["private_key"] = resp.Data.Config.PrivateKey
		c["passphrase"] = resp.Data.Config.Passphrase
	
		if strToBool(resp.Data.Config.IsPrivateKeyEncrypted) {
			// we should ignore default `false` value if not configured to prevent data drifts
			// we read it only if `true` to prevent false data drifts
			c["is_private_key_encrypted"] = resp.Data.Config.IsPrivateKeyEncrypted
		}
	}

	c["connection_type"] = resourceDestinationConfigNormalizeConnectionType(resp.Data.Config.ConnectionType)
	c["tunnel_host"] = resp.Data.Config.TunnelHost
	c["tunnel_port"] = resp.Data.Config.TunnelPort
	c["tunnel_user"] = resp.Data.Config.TunnelUser
	c["project_id"] = resp.Data.Config.ProjectID

	// BQ returns its data_set_location as location in response
	if resp.Data.Config.Location != "" && resourceDestinationIsBigQuery(resp.Data.Service) {
		c["data_set_location"] = resp.Data.Config.Location
	} else {
		c["data_set_location"] = resp.Data.Config.DataSetLocation
	}

	c["bucket"] = resp.Data.Config.Bucket
	c["server_host_name"] = resp.Data.Config.ServerHostName
	c["http_path"] = resp.Data.Config.HTTPPath
	c["create_external_tables"] = resp.Data.Config.CreateExternalTables
	c["external_location"] = resp.Data.Config.ExternalLocation
	c["auth_type"] = resp.Data.Config.AuthType
	c["cluster_id"] = resp.Data.Config.ClusterId
	c["cluster_region"] = resp.Data.Config.ClusterRegion
	c["public_key"] = resp.Data.Config.PublicKey
	c["role"] = resp.Data.Config.Role
	c["catalog"] = resp.Data.Config.Catalog
	c["fivetran_role_arn"] = resp.Data.Config.FivetranRoleArn
	c["prefix_path"] = resp.Data.Config.PrefixPath
	c["region"] = resp.Data.Config.Region

	config = append(config, c)

	return config, nil
}

func resourceDestinationIsBigQuery(service string) bool {
	return service == "big_query" || service == "managed_big_query" || service == "big_query_dts"
}

// resourceDestinationCreateConfig receives a config type []interface{} and returns a
// *fivetran.DestinationConfig and a ok value. The ok value is true if any configuration
// has been set.
func resourceDestinationCreateConfig(config []interface{}) (*destinations.DestinationConfig, bool) {
	fivetranConfig := fivetran.NewDestinationConfig()
	var hasConfig bool

	c := config[0].(map[string]interface{})

	if v := c["create_external_tables"].(string); v != "" {
		fivetranConfig.CreateExternalTables(strToBool(v))
		hasConfig = true
	}
	if v := c["host"].(string); v != "" {
		fivetranConfig.Host(v)
		hasConfig = true
	}
	if v := c["port"].(int); v != 0 {
		fivetranConfig.Port(v)
		hasConfig = true
	}
	if v := c["database"].(string); v != "" {
		fivetranConfig.Database(v)
		hasConfig = true
	}
	if v := c["auth"].(string); v != "" {
		fivetranConfig.Auth(v)
		hasConfig = true
	}
	if v := c["user"].(string); v != "" {
		fivetranConfig.User(v)
		hasConfig = true
	}
	if v := c["password"].(string); v != "" {
		fivetranConfig.Password(v)
		hasConfig = true
	}
	if v := c["connection_type"].(string); v != "" {
		fivetranConfig.ConnectionType(v)
		hasConfig = true
	}
	if v := c["tunnel_host"].(string); v != "" {
		fivetranConfig.TunnelHost(v)
		hasConfig = true
	}
	if v := c["tunnel_port"].(string); v != "" {
		fivetranConfig.TunnelPort(v)
		hasConfig = true
	}
	if v := c["tunnel_user"].(string); v != "" {
		fivetranConfig.TunnelUser(v)
		hasConfig = true
	}
	if v := c["project_id"].(string); v != "" {
		fivetranConfig.ProjectID(v)
		hasConfig = true
	}
	if v := c["data_set_location"].(string); v != "" {
		fivetranConfig.DataSetLocation(v)
		hasConfig = true
	}
	if v := c["bucket"].(string); v != "" {
		fivetranConfig.Bucket(v)
		hasConfig = true
	}
	if v := c["server_host_name"].(string); v != "" {
		fivetranConfig.ServerHostName(v)
		hasConfig = true
	}
	if v := c["http_path"].(string); v != "" {
		fivetranConfig.HTTPPath(v)
		hasConfig = true
	}
	if v := c["personal_access_token"].(string); v != "" {
		fivetranConfig.PersonalAccessToken(v)
		hasConfig = true
	}
	if v := c["external_location"].(string); v != "" {
		fivetranConfig.ExternalLocation(v)
		hasConfig = true
	}
	if v := c["auth_type"].(string); v != "" {
		fivetranConfig.AuthType(v)
		hasConfig = true
	}
	if v := c["role_arn"].(string); v != "" {
		fivetranConfig.RoleArn(v)
		hasConfig = true
	}
	if v := c["secret_key"].(string); v != "" {
		fivetranConfig.SecretKey(v)
		hasConfig = true
	}
	if v := c["private_key"].(string); v != "" {
		fivetranConfig.PrivateKey(v)
		hasConfig = true
	}
	if v := c["cluster_id"].(string); v != "" {
		fivetranConfig.ClusterId(v)
		hasConfig = true
	}
	if v := c["cluster_region"].(string); v != "" {
		fivetranConfig.ClusterRegion(v)
		hasConfig = true
	}
	if v := c["role"].(string); v != "" {
		fivetranConfig.Role(v)
		hasConfig = true
	}
	if v := c["is_private_key_encrypted"].(string); v != "" {
		fivetranConfig.IsPrivateKeyEncrypted(strToBool(v))
		hasConfig = true
	}
	if v := c["passphrase"].(string); v != "" {
		fivetranConfig.Passphrase(v)
		hasConfig = true
	}
	if v := c["catalog"].(string); v != "" {
		fivetranConfig.Catalog(v)
		hasConfig = true
	}
	if v := c["fivetran_role_arn"].(string); v != "" {
		fivetranConfig.FivetranRoleArn(v)
		hasConfig = true
	}
	if v := c["prefix_path"].(string); v != "" {
		fivetranConfig.PrefixPath(v)
		hasConfig = true
	}
	if v := c["region"].(string); v != "" {
		fivetranConfig.Region(v)
		hasConfig = true
	}

	return fivetranConfig, hasConfig
}

func resourceDestinationConfigNormalizeConnectionType(connectionType string) string {
	if connectionType == "SshTunnel" {
		return "SSHTunnel"
	}
	return connectionType
}
