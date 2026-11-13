package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ExternalLoggingV2ResourceSchema() resourceSchema.Schema {
	return resourceSchema.Schema{
		Description: "Manages the external logging service of a destination (group).",
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the log service within the Fivetran system. It matches `group_id`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_id": resourceSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the group (destination) the log service belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				Description: "Service-specific log service configuration. The accepted fields depend on `service`; when the Fivetran metadata API describes the log service they are validated at plan time. `config` is not sensitive: set secrets in `config_secrets`, or pass them from sensitive variables.",
			},
			"config_secrets": resourceSchema.DynamicAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: "Write-only secret config fields (fields with `format: password` in log service metadata, for example `api_key` or `token`). Accepts ephemeral values and is never stored in state. Values are sent on create and whenever `config_secrets_version` changes. Requires Terraform 1.11 or later.",
			},
			"config_secrets_version": resourceSchema.Int64Attribute{
				Optional:    true,
				Description: "Change this value to send the current `config_secrets` to Fivetran, for example after rotating a secret.",
			},
		},
	}
}
