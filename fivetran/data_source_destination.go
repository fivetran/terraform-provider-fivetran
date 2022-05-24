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
			"id":               {Type: schema.TypeString, Required: true},
			"group_id":         {Type: schema.TypeString, Computed: true},
			"service":          {Type: schema.TypeString, Computed: true},
			"region":           {Type: schema.TypeString, Computed: true},
			"time_zone_offset": {Type: schema.TypeString, Computed: true},
			"config":           dataSourceDestinationSchemaConfig(),
			"setup_status":     {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataSourceDestinationSchemaConfig() *schema.Schema {
	return &schema.Schema{Type: schema.TypeSet, Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"host":                   {Type: schema.TypeString, Computed: true},
				"port":                   {Type: schema.TypeInt, Computed: true},
				"database":               {Type: schema.TypeString, Computed: true},
				"auth":                   {Type: schema.TypeString, Computed: true},
				"user":                   {Type: schema.TypeString, Computed: true},
				"password":               {Type: schema.TypeString, Computed: true},
				"connection_type":        {Type: schema.TypeString, Computed: true},
				"tunnel_host":            {Type: schema.TypeString, Computed: true},
				"tunnel_port":            {Type: schema.TypeString, Computed: true},
				"tunnel_user":            {Type: schema.TypeString, Computed: true},
				"project_id":             {Type: schema.TypeString, Computed: true},
				"data_set_location":      {Type: schema.TypeString, Computed: true},
				"bucket":                 {Type: schema.TypeString, Computed: true},
				"server_host_name":       {Type: schema.TypeString, Computed: true},
				"http_path":              {Type: schema.TypeString, Computed: true},
				"personal_access_token":  {Type: schema.TypeString, Computed: true},
				"create_external_tables": {Type: schema.TypeString, Computed: true},
				"external_location":      {Type: schema.TypeString, Computed: true},
				"auth_type":              {Type: schema.TypeString, Computed: true},
				"role_arn":               {Type: schema.TypeString, Computed: true},
				"secret_key":             {Type: schema.TypeString, Computed: true},
				"cluster_id":             {Type: schema.TypeString, Computed: true},
				"cluster_region":         {Type: schema.TypeString, Computed: true},
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
	c["data_set_location"] = resp.Data.Config.DataSetLocation
	c["bucket"] = resp.Data.Config.Bucket
	c["server_host_name"] = resp.Data.Config.ServerHostName
	c["http_path"] = resp.Data.Config.HTTPPath
	c["personal_access_token"] = resp.Data.Config.PersonalAccessToken
	c["create_external_tables"] = resp.Data.Config.CreateExternalTables
	c["external_location"] = resp.Data.Config.ExternalLocation
	c["auth_type"] = resp.Data.Config.AuthType
	c["role_arn"] = resp.Data.Config.RoleArn
	c["secret_key"] = resp.Data.Config.SecretKey
	c["cluster_id"] = resp.Data.Config.ClusterId
	c["cluster_region"] = resp.Data.Config.ClusterRegion
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
