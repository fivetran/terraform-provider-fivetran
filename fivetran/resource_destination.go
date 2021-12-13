package fivetran

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDestinationCreate,
		ReadContext:   resourceDestinationRead,
		UpdateContext: resourceDestinationUpdate,
		DeleteContext: resourceDestinationDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id":                 {Type: schema.TypeString, Computed: true},
			"group_id":           {Type: schema.TypeString, Required: true, ForceNew: true},
			"service":            {Type: schema.TypeString, Required: true, ForceNew: true},
			"region":             {Type: schema.TypeString, Required: true},
			"time_zone_offset":   {Type: schema.TypeString, Required: true},
			"config":             resourceDestinationSchemaConfig(),
			"trust_certificates": {Type: schema.TypeBool, Optional: true, ForceNew: true}, // T-112419, ForceNew can be removed and the field can be updated
			"trust_fingerprints": {Type: schema.TypeBool, Optional: true, ForceNew: true}, // T-112419, ForceNew can be removed and the field can be updated
			// "run_setup_tests" default value is true. It is set as required to avoid confusion, so the expected behaviour should
			// be explicitly declared.
			"run_setup_tests": {Type: schema.TypeBool, Required: true, ForceNew: true}, // T-112419, ForceNew can be removed and the field can be updated
			"setup_status":    {Type: schema.TypeString, Computed: true},
			// "setup_tests": ... // missing /T-112419
			"last_updated": {Type: schema.TypeString, Computed: true}, // internal
		},
	}
}

func resourceDestinationSchemaConfig() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Required: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"host":     {Type: schema.TypeString, Optional: true},
				"port":     {Type: schema.TypeInt, Optional: true},
				"database": {Type: schema.TypeString, Optional: true},
				"auth":     {Type: schema.TypeString, Optional: true},
				"user":     {Type: schema.TypeString, Optional: true},
				"password": {Type: schema.TypeString, Optional: true, Sensitive: true},
				// "connection_type", for example, is mandatory for some destinations. Configurations settings
				// should be explicitly declared. Doing that, we avoid not declaring "connection_type"
				// (or any other configuration) and the REST API returning it.
				"connection_type":        {Type: schema.TypeString, Optional: true},
				"tunnel_host":            {Type: schema.TypeString, Optional: true},
				"tunnel_port":            {Type: schema.TypeString, Optional: true},
				"tunnel_user":            {Type: schema.TypeString, Optional: true},
				"project_id":             {Type: schema.TypeString, Optional: true},
				"data_set_location":      {Type: schema.TypeString, Optional: true},
				"bucket":                 {Type: schema.TypeString, Optional: true},
				"server_host_name":       {Type: schema.TypeString, Optional: true},
				"http_path":              {Type: schema.TypeString, Optional: true},
				"personal_access_token":  {Type: schema.TypeString, Optional: true, Sensitive: true},
				"create_external_tables": {Type: schema.TypeString, Optional: true},
				"external_location":      {Type: schema.TypeString, Optional: true},
				"auth_type":              {Type: schema.TypeString, Optional: true},
				"role_arn":               {Type: schema.TypeString, Optional: true, Sensitive: true},
				"secret_key":             {Type: schema.TypeString, Optional: true, Sensitive: true},
				"cluster_id":             {Type: schema.TypeString, Computed: true},
				"cluster_region":         {Type: schema.TypeString, Computed: true},
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
	svc.RunSetupTests(d.Get("run_setup_tests").(bool))

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

	return diags
}

func resourceDestinationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDestinationModify()

	svc.DestinationID(d.Get("id").(string))

	if d.HasChange("region") {
		svc.Region(d.Get("region").(string))
	}
	if d.HasChange("time_zone_offset") {
		svc.TimeZoneOffset(d.Get("time_zone_offset").(string))
	}
	if d.HasChange("config") {
		_, n := d.GetChange("config")
		// resourceDestinationCreateConfig is used here because
		// the whole "config" block must be sent to the REST API.
		if v, ok := resourceDestinationCreateConfig(n.([]interface{})); ok {
			svc.Config(v)
			// only sets change if func resourceDestinationCreateConfig returns ok
		}
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
func resourceDestinationReadConfig(resp *fivetran.DestinationDetailsResponse, currentConfig []interface{}) ([]interface{}, error) {
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
	// The REST API sends the password field masked. We use the state stored password here.
	c["password"] = currentConfig[0].(map[string]interface{})["password"].(string)
	// connection_type is returned as ConnectionMethod
	c["connection_type"] = dataSourceDestinationConfigNormalizeConnectionType(resp.Data.Config.ConnectionMethod)
	c["tunnel_host"] = resp.Data.Config.TunnelHost
	c["tunnel_port"] = resp.Data.Config.TunnelPort
	c["tunnel_user"] = resp.Data.Config.TunnelUser
	c["project_id"] = resp.Data.Config.ProjectID
	c["data_set_location"] = resp.Data.Config.DataSetLocation
	c["bucket"] = resp.Data.Config.Bucket
	c["server_host_name"] = resp.Data.Config.ServerHostName
	c["http_path"] = resp.Data.Config.HTTPPath
	c["personal_access_token"] = resp.Data.Config.PersonalAccessToken
	if resp.Data.Config.CreateExternalTables != nil {
		c["create_external_tables"] = boolPointerToStr(resp.Data.Config.CreateExternalTables)
	}
	c["external_location"] = resp.Data.Config.ExternalLocation
	c["auth_type"] = resp.Data.Config.AuthType
	c["role_arn"] = resp.Data.Config.RoleArn
	c["secret_key"] = resp.Data.Config.SecretKey
	c["cluster_id"] = resp.Data.Config.ClusterId
	c["cluster_region"] = resp.Data.Config.ClusterRegion

	config = append(config, c)

	return config, nil
}

// resourceDestinationCreateConfig receives a config type []interface{} and returns a
// *fivetran.DestinationConfig and a ok value. The ok value is true if any configuration
// has been set.
func resourceDestinationCreateConfig(config []interface{}) (*fivetran.DestinationConfig, bool) {
	fivetranConfig := fivetran.NewDestinationConfig()
	var hasConfig bool

	if v := config[0].(map[string]interface{})["create_external_tables"].(string); v != "" {
		fivetranConfig.CreateExternalTables(strToBool(v))
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["host"].(string); v != "" {
		fivetranConfig.Host(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["port"].(int); v != 0 {
		fivetranConfig.Port(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["database"].(string); v != "" {
		fivetranConfig.Database(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["auth"].(string); v != "" {
		fivetranConfig.Auth(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["user"].(string); v != "" {
		fivetranConfig.User(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["password"].(string); v != "" {
		fivetranConfig.Password(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["connection_type"].(string); v != "" {
		fivetranConfig.ConnectionType(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["tunnel_host"].(string); v != "" {
		fivetranConfig.TunnelHost(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["tunnel_port"].(string); v != "" {
		fivetranConfig.TunnelPort(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["tunnel_user"].(string); v != "" {
		fivetranConfig.TunnelUser(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["project_id"].(string); v != "" {
		fivetranConfig.ProjectID(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["data_set_location"].(string); v != "" {
		fivetranConfig.DataSetLocation(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["bucket"].(string); v != "" {
		fivetranConfig.Bucket(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["server_host_name"].(string); v != "" {
		fivetranConfig.ServerHostName(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["http_path"].(string); v != "" {
		fivetranConfig.HTTPPath(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["personal_access_token"].(string); v != "" {
		fivetranConfig.PersonalAccessToken(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["external_location"].(string); v != "" {
		fivetranConfig.ExternalLocation(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["auth_type"].(string); v != "" {
		fivetranConfig.AuthType(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["role_arn"].(string); v != "" {
		fivetranConfig.RoleArn(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["secret_key"].(string); v != "" {
		fivetranConfig.SecretKey(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["cluster_id"].(string); v != "" {
		fivetranConfig.ClusterId(v)
		hasConfig = true
	}
	if v := config[0].(map[string]interface{})["cluster_region"].(string); v != "" {
		fivetranConfig.ClusterRegion(v)
		hasConfig = true
	}

	return fivetranConfig, hasConfig
}
