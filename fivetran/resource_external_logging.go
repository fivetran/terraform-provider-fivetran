package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceExternalLogging() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExternalLoggingCreate,
		ReadContext:   resourceExternalLoggingRead,
		UpdateContext: resourceExternalLoggingUpdate,
		DeleteContext: resourceExternalLoggingDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: 	   getExternalLoggingSchema(false),
	}
}

func getExternalLoggingSchema(datasource bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    !datasource,
                Required:    datasource,
				Description: "The unique identifier for the log service within the Fivetran system.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Computed:    datasource,
				Required:    !datasource,
				ForceNew:    !datasource,
				Description: "The unique identifier for the log service within the Fivetran system.",
			},
			"service": {
				Type:        schema.TypeString,
				Required:    !datasource,
				Computed:    datasource,
				ForceNew:    true,
				Description: "The name for the log service type within the Fivetran system. We support the following log services: azure_monitor_log, cloudwatch, datadog_log, new_relic_log, splunkLog, stackdriver.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    !datasource,
				Computed:    datasource,
				Description: "The boolean value specifying whether the log service is enabled.",
			},
			"run_setup_tests": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies whether the setup tests should be run automatically. The default value is TRUE.",
			},
			"config": resourceExternalLoggingSchemaConfig(datasource),
		}
}

func resourceExternalLoggingSchemaConfig(datasource bool) *schema.Schema {
	maxItems := 1
	if datasource {
		maxItems = 0
	}

	return &schema.Schema{
		Type: schema.TypeList, 
		Required: !datasource,
		Optional: datasource,
        Computed: datasource,
		MaxItems: maxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"workspace_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Workspace ID",
				},
				"primary_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "Primary Key",
				},
				"log_group_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Log Group Name",
				},
				"role_arn": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Role Arn",
				},
				"external_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "external_id",
				},
				"region": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Region",
				},
				"api_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "API Key",
				},
				"sub_domain": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Sub Domain",
				},
				"host": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Server name",
				},
				"hostname": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Server name",
				},
				"port": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "Port",
				},
				"channel": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Channel",
				},
				"enable_ssl": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Enable SSL",
				},
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "Token",
				},
			},
		},
	}
}

func resourceExternalLoggingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewExternalLoggingCreate()

	svc.GroupId(d.Get("group_id").(string))
	svc.Service(d.Get("service").(string))
	if v, ok := d.GetOk("enabled"); ok {
		svc.Enabled(v.(bool))
	}
	if v, ok := resourceExternalLoggingCreateConfig(d.Get("config").([]interface{})); ok {
		svc.Config(v)
	}

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.Id)
	resourceExternalLoggingRead(ctx, d, m)

	return diags
}

func resourceExternalLoggingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewExternalLoggingDetails()

	resp, err := svc.ExternalLoggingId(d.Get("id").(string)).Do(ctx)
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
	mapStringInterface := make(map[string]interface{})
	mapAddStr(mapStringInterface, "id", resp.Data.Id)
	mapAddStr(mapStringInterface, "service", resp.Data.Service)
	mapStringInterface["enabled"] = resp.Data.Enabled

	config, err := resourceExternalLoggingReadConfig(&resp, d.Get("config").([]interface{}))
	if err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}
	mapStringInterface["config"] = config

	for k, v := range mapStringInterface {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.Id)

	return diags
}

func resourceExternalLoggingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewExternalLoggingModify()

	svc.ExternalLoggingId(d.Get("id").(string))
	hasChanges := false

	if d.HasChange("enabled") {
		svc.Enabled(d.Get("enabled").(bool))
		hasChanges = true
	}

	if d.HasChange("config") {
		_, n := d.GetChange("config")
		// resourceExternalLoggingCreateConfig is used here because
		// the whole "config" block must be sent to the REST API.
		if v, ok := resourceExternalLoggingCreateConfig(n.([]interface{})); ok {
			svc.Config(v)
			hasChanges = true
			// only sets change if func resourceExternalLoggingCreateConfig returns ok
		}
	}
	if hasChanges {
		if v, ok := d.GetOk("run_setup_tests"); ok {
			svc.RunSetupTests(v.(bool))
		}

		resp, err := svc.Do(ctx)
		if err != nil {
			// resourceExternalLoggingRead here makes sure the state is updated after a NewExternalLoggingModify error.
			diags = resourceExternalLoggingRead(ctx, d, m)
			return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
		}
	} else {
		// if only "run_setup_tests" updated to true - setup tests should be performed without update request
		if v, ok := d.GetOk("run_setup_tests"); ok && v.(bool) && d.HasChange("run_setup_tests") {
			testsSvc := client.NewExternalLoggingSetupTests().ExternalLoggingId(d.Get("id").(string))
			resp, err := testsSvc.Do(ctx)
			if err != nil {
				return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
			}
		}
	}

	return resourceExternalLoggingRead(ctx, d, m)
}

func resourceExternalLoggingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewExternalLoggingDelete()

	resp, err := svc.ExternalLoggingId(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}

// resourceExternalLoggingCreateConfig receives a config type []interface{} and returns a
// *fivetran.ExternalLoggingConfig and a ok value. The ok value is true if any configuration
// has been set.
func resourceExternalLoggingCreateConfig(config []interface{}) (*fivetran.ExternalLoggingConfig, bool) {
	fivetranConfig := fivetran.NewExternalLoggingConfig()
	var hasConfig bool

	c := config[0].(map[string]interface{})

	if v := c["workspace_id"].(string); v != "" {
		fivetranConfig.WorkspaceId(v)
		hasConfig = true
	}
	if v := c["port"].(int); v != 0 {
		fivetranConfig.Port(v)
		hasConfig = true
	}
	if v := c["log_group_name"].(string); v != "" {
		fivetranConfig.LogGroupName(v)
		hasConfig = true
	}
	if v := c["role_arn"].(string); v != "" {
		fivetranConfig.RoleArn(v)
		hasConfig = true
	}
	if v := c["external_id"].(string); v != "" {
		fivetranConfig.ExternalId(v)
		hasConfig = true
	}
	if v := c["region"].(string); v != "" {
		fivetranConfig.Region(v)
		hasConfig = true
	}
	if v := c["sub_domain"].(string); v != "" {
		fivetranConfig.SubDomain(v)
		hasConfig = true
	}
	if v := c["host"].(string); v != "" {
		fivetranConfig.Host(v)
		hasConfig = true
	}
	if v := c["hostname"].(string); v != "" {
		fivetranConfig.Hostname(v)
		hasConfig = true
	}
	if v := c["enable_ssl"].(bool); v != "" {
		fivetranConfig.EnableSsl(v) // here will be a bug currently - we need to fix go-fivetran client. Or just use custom config.
		hasConfig = true
	}
	if v := c["channel"].(string); v != "" {
		fivetranConfig.Channel(v)
		hasConfig = true
	}
	if v := c["primary_key"].(string); v != "" {
		fivetranConfig.PrimaryKey(v)
		hasConfig = true
	}
	if v := c["api_key"].(string); v != "" {
		fivetranConfig.ApiKey(v)
		hasConfig = true
	}
	if v := c["token"].(string); v != "" {
		fivetranConfig.Token(v)
		hasConfig = true
	}

	return fivetranConfig, hasConfig
}


func resourceExternalLoggingReadConfig(resp *fivetran.ExternalLoggingDetailsResponse, currentConfig []interface{}) ([]interface{}, error) {
	var config []interface{}

	c := make(map[string]interface{})
	c["workspace_id"] = resp.Data.Config.WorkspaceId
	c["port"] = resp.Data.Config.Port
	c["log_group_name"] = resp.Data.Config.LogGroupName
	c["role_arn"] = resp.Data.Config.RoleArn
	c["external_id"] = resp.Data.Config.ExternalId
	c["region"] = resp.Data.Config.Region
	c["sub_domain"] = resp.Data.Config.SubDomain
	c["host"] = resp.Data.Config.Host
	c["hostname"] = resp.Data.Config.Hostname
	c["enable_ssl"] = resp.Data.Config.EnableSsl
	c["channel"] = resp.Data.Config.Channel
			
	if len(currentConfig) > 0 {
		// The REST API sends the password field masked. We use the state stored password here if possible.
		currentConfigMap := currentConfig[0].(map[string]interface{})
		c["primary_key"] = currentConfigMap["primary_key"]
		c["api_key"] = currentConfigMap["api_key"].(string)
		c["token"] = currentConfigMap["token"].(string)
	} else {
		c["primary_key"] = resp.Data.Config.PrimaryKey
    	c["api_key"] = resp.Data.Config.ApiKey
    	c["token"] = resp.Data.Config.Token
	}

	config = append(config, c)

	return config, nil
}