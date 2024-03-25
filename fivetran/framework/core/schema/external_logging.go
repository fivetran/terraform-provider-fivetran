package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
)

func ExternalLoggingResource() resourceSchema.Schema {
	return resourceSchema.Schema {
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the log service within the Fivetran system.",
			},
			"group_id": resourceSchema.StringAttribute{
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "The unique identifier for the log service within the Fivetran system.",
			},
			"service": resourceSchema.StringAttribute{
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "The name for the log service type within the Fivetran system. We support the following log services: azure_monitor_log, cloudwatch, datadog_log, new_relic_log, splunkLog, stackdriver.",
			},
			"enabled": resourceSchema.BoolAttribute{
				Optional:	 true,
				Description: "The boolean value specifying whether the log service is enabled.",
			},
			"run_setup_tests": resourceSchema.BoolAttribute{
				Optional:	 true,
				Description: "Specifies whether the setup tests should be run automatically. The default value is TRUE.",
			},
		},
		Blocks: map[string]resourceSchema.Block{
			"config": resourceSchema.SingleNestedBlock{
				Attributes: map[string]resourceSchema.Attribute{
					"workspace_id": resourceSchema.StringAttribute{ 
						Optional:    true,
						Description: "Workspace ID",
					},
					"primary_key": resourceSchema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Primary Key",
					},
					"log_group_name": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "Log Group Name",
					},
					"role_arn": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "Role Arn",
					},
					"external_id": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "external_id",
					},
					"region": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "Region",
					},
					"api_key": resourceSchema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "API Key",
					},
					"sub_domain": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "Sub Domain",
					},
					"host": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "Server name",
					},
					"hostname": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "Server name",
					},
					"port": resourceSchema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Port",
					},
					"channel": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "Channel",
					},
					"enable_ssl": resourceSchema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Enable SSL",
					},
					"token": resourceSchema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Token",
					},
					"project_id": resourceSchema.StringAttribute{
						Optional:    true,
						Description: "Project Id for Google Cloud Logging",
					},
				},
			},
		},
	}
}


func ExternalLoggingDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the log service within the Fivetran system.",
			},
			"group_id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the log service within the Fivetran system.",
			},
			"service": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The name for the log service type within the Fivetran system. We support the following log services: azure_monitor_log, cloudwatch, datadog_log, new_relic_log, splunkLog, stackdriver.",
			},
			"enabled": datasourceSchema.BoolAttribute{
				Computed:	 true,
				Description: "The boolean value specifying whether the log service is enabled.",
			},
			"run_setup_tests": datasourceSchema.BoolAttribute{
				Optional:	 true,
				Description: "Specifies whether the setup tests should be run automatically. The default value is TRUE.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"config": datasourceSchema.SingleNestedBlock{
				Attributes: map[string]datasourceSchema.Attribute{
					"workspace_id": datasourceSchema.StringAttribute{ 
						Optional:    true,
						Description: "Workspace ID",
					},
					"primary_key": datasourceSchema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Primary Key",
					},
					"log_group_name": datasourceSchema.StringAttribute{
						Optional:    true,
						Description: "Log Group Name",
					},
					"role_arn": datasourceSchema.StringAttribute{
						Optional:    true,
						Description: "Role Arn",
					},
					"external_id": datasourceSchema.StringAttribute{
						Optional:    true,
						Description: "external_id",
					},
					"region": datasourceSchema.StringAttribute{
						Optional:    true,
						Description: "Region",
					},
					"api_key": datasourceSchema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "API Key",
					},
					"sub_domain": datasourceSchema.StringAttribute{
						Optional:    true,
						Description: "Sub Domain",
					},
					"host": datasourceSchema.StringAttribute{
						Optional:    true,
						Description: "Server name",
					},
					"hostname": datasourceSchema.StringAttribute{
						Optional:    true,
						Description: "Server name",
					},
					"port": datasourceSchema.Int64Attribute{
						Optional:    true,
						Description: "Port",
					},
					"channel": datasourceSchema.StringAttribute{
						Optional:    true,
						Description: "Channel",
					},
					"enable_ssl": datasourceSchema.BoolAttribute{
						Optional:    true,
						Description: "Enable SSL",
					},
					"token": datasourceSchema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Token",
					},
					"project_id": datasourceSchema.StringAttribute{
						Optional:    true,
						Description: "Project Id for Google Cloud Logging",
					},
				},
			},
		},
	}
}