package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func AccountExternalLoggingV2ResourceSchema() resourceSchema.Schema {
	return resourceSchema.Schema{
		Description: "Manages the account-level external logging service.",
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the account-level log service within the Fivetran system.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service": resourceSchema.StringAttribute{
				Required:    true,
				Description: "The name of the log service type within the Fivetran system. Supported log services: `azure_monitor_log`, `cloudwatch`, `datadog_log`, `dynatrace`, `grafana_loki`, `new_relic_log`, `splunkLog`, `stackdriver`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": resourceSchema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "The boolean value specifying whether the log service is enabled. Defaults to `true`.",
			},
			"run_setup_tests": resourceSchema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Specifies whether the setup tests should be run after the log service is created or updated. Defaults to `false`. This is a plan-only attribute and is not read back from the API.",
			},
			"config": resourceSchema.DynamicAttribute{
				Required:    true,
				Description: "Service-specific log service configuration. The accepted fields depend on `service`; when the Fivetran metadata API describes the log service they are validated at plan time.",
			},
		},
	}
}
