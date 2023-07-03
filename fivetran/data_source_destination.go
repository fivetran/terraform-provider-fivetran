package fivetran

import (
	"context"
	"fmt"
	"strconv"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDestination() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDestinationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier for the destination within the Fivetran system",
			},
			"group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier for the Group within the Fivetran system.",
			},
			"service": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The connector type name within the Fivetran system",
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
			},
			"time_zone_offset": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Determines the time zone for the Fivetran sync schedule.",
			},
			"config": dataSourceDestinationSchemaConfig(),
			"setup_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Destination setup status",
			},
		},
	}
}

func dataSourceDestinationSchemaConfig() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeSet,
		// Uncomment Optional:true, before re-generating docs
		//Optional: true,
		Computed: true,
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

func dataSourceDestinationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDestinationDetails()

	resp, err := svc.DestinationID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.ID
	msi["group_id"] = resp.Data.GroupID
	msi["service"] = resp.Data.Service
	msi["region"] = resp.Data.Region
	msi["time_zone_offset"] = resp.Data.TimeZoneOffset
	config, err := dataSourceDestinationConfig(&resp)
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

// dataSourceDestinationConfig receives a *fivetran.DestinationDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" set.
func dataSourceDestinationConfig(resp *fivetran.DestinationDetailsResponse) ([]interface{}, error) {
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
	c["password"] = resp.Data.Config.Password
	c["connection_type"] = dataSourceDestinationConfigNormalizeConnectionType(resp.Data.Config.ConnectionType)
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
	c["personal_access_token"] = resp.Data.Config.PersonalAccessToken
	c["create_external_tables"] = resp.Data.Config.CreateExternalTables
	c["external_location"] = resp.Data.Config.ExternalLocation
	c["auth_type"] = resp.Data.Config.AuthType
	c["role_arn"] = resp.Data.Config.RoleArn
	c["secret_key"] = resp.Data.Config.SecretKey
	c["private_key"] = resp.Data.Config.PrivateKey
	c["public_key"] = resp.Data.Config.PublicKey
	c["cluster_id"] = resp.Data.Config.ClusterId
	c["cluster_region"] = resp.Data.Config.ClusterRegion
	c["role"] = resp.Data.Config.Role
	c["is_private_key_encrypted"] = resp.Data.Config.IsPrivateKeyEncrypted
	c["passphrase"] = resp.Data.Config.Passphrase
	c["catalog"] = resp.Data.Config.Catalog
	c["fivetran_role_arn"] = resp.Data.Config.FivetranRoleArn
	c["prefix_path"] = resp.Data.Config.PrefixPath
	c["region"] = resp.Data.Config.Region

	config = append(config, c)

	return config, nil
}

// dataSourceDestinationConfigNormalizeConnectionType normalizes *fivetran.DestinationDetailsResponse.Data.Config.ConnectionType. /T-111758.
func dataSourceDestinationConfigNormalizeConnectionType(connectionType string) string {
	if connectionType == "SshTunnel" {
		return "SSHTunnel"
	}
	return connectionType
}
