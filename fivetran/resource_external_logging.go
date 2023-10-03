package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	externallogging "github.com/fivetran/go-fivetran/external_logging"
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
		Schema:        getExternalLoggingSchema(false),
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
		Type:     schema.TypeList,
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
					Type:        schema.TypeBool,
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
	svc.Enabled(d.Get("enabled").(bool))
	config := resourceExternalLoggingCreateConfig(d)

	svc.ConfigCustom(&config)

	resp, err := svc.DoCustom(ctx)
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
		config := resourceExternalLoggingCreateConfig(d)
		svc.ConfigCustom(&config)

		hasChanges = true
	}
	if hasChanges {
		if v, ok := d.GetOk("run_setup_tests"); ok {
			svc.RunSetupTests(v.(bool))
		}

		resp, err := svc.DoCustom(ctx)
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

func resourceExternalLoggingCreateConfig(d *schema.ResourceData) map[string]interface{} {
	configMap := make(map[string]interface{})

	var config = d.Get("config").([]interface{})

	if len(config) < 1 || config[0] == nil {
		return configMap
	}

	c := config[0].(map[string]interface{})

	return c
}

func resourceExternalLoggingReadConfig(resp *externallogging.ExternalLoggingResponse, currentConfig []interface{}) ([]interface{}, error) {
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
