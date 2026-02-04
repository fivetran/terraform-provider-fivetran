package schema

import (
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ExternalLoggingResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the log service within the Fivetran system.",
			},
			"group_id": resourceSchema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The unique identifier for the log service within the Fivetran system.",
			},
			"service": resourceSchema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The name for the log service type within the Fivetran system. We support the following log services: azure_monitor_log, cloudwatch, datadog_log, new_relic_log, splunkLog, stackdriver.",
			},
			"enabled": resourceSchema.BoolAttribute{
				Optional:    true,
				Description: "The boolean value specifying whether the log service is enabled.",
			},
			"run_setup_tests": resourceSchema.BoolAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
				Description:   "Specifies whether the setup tests should be run automatically. The default value is TRUE.",
			},
		},
		Blocks: map[string]resourceSchema.Block{
			"config": resourceSchema.SingleNestedBlock{
				Attributes: GetResourceExternalLoggingConfigSchemaAttributes(),
			},
		},
	}
}

func ExternalLoggingDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
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
				Computed:    true,
				Description: "The boolean value specifying whether the log service is enabled.",
			},
			"run_setup_tests": datasourceSchema.BoolAttribute{
				Optional:    true,
				Description: "Specifies whether the setup tests should be run automatically. The default value is TRUE.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"config": datasourceSchema.SingleNestedBlock{
				Attributes: GetDatasourceExternalLoggingConfigSchemaAttributes(),
			},
		},
	}
}

func ExternalLogsDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"logs": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the log service within the Fivetran system.",
						},
						"service": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The name for the log service type within the Fivetran system. We support the following log services: azure_monitor_log, cloudwatch, datadog_log, new_relic_log, splunkLog, stackdriver.",
						},
						"enabled": datasourceSchema.BoolAttribute{
							Computed:    true,
							Description: "The boolean value specifying whether the log service is enabled.",
						},
					},
				},
			},
		},
	}
}
