package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func DestinationV2ResourceSchema() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: DestinationV2ResourceAttributes(),
		Version:    0,
	}
}

func DestinationV2ResourceAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			Computed:    true,
			Description: "The unique identifier for the destination within the Fivetran system.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"group_id": resourceSchema.StringAttribute{
			Required:    true,
			Description: "The unique identifier for the Group within the Fivetran system.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"service": resourceSchema.StringAttribute{
			Required:    true,
			Description: "The destination service type (e.g., `snowflake`, `big_query`, `databricks`). See Fivetran destination types documentation for available services.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"config": resourceSchema.DynamicAttribute{
			Required:    true,
			Description: "Service-specific destination configuration. The accepted fields are defined by destination metadata at runtime.",
		},
		"region": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Data processing location. This is where Fivetran will operate and run computation on data. If not set, the API default region is used.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"time_zone_offset": resourceSchema.StringAttribute{
			Required:    true,
			Description: "Determines the time zone for the Fivetran sync schedule. Possible values: -11 to +12.",
		},
		"daylight_saving_time_enabled": resourceSchema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Shift the sync schedule to account for daylight saving time (for time_zone_offset values that observe DST).",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"setup_status": resourceSchema.StringAttribute{
			Computed:    true,
			Description: "Destination setup status. Possible values: `incomplete`, `connected`, `broken`.",
		},
		"networking_method": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The networking method for the destination. Possible values: `Directly`, `SshTunnel`, `ProxyAgent`, `PrivateLink`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"private_link_id": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The private link ID. Required when `networking_method` is `PrivateLink` and the destination supports self-service PrivateLink.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"proxy_agent_id": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The ID of the proxy agent to use. Required when `networking_method` is `ProxyAgent`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"hybrid_deployment_agent_id": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The hybrid deployment agent ID that refers to the controller created for the group the destination belongs to.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"run_setup_tests": resourceSchema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether to run setup tests when creating or updating the destination. Defaults to `true`, matching the API.",
		},
		"trust_certificates": resourceSchema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Specifies whether Fivetran should trust certificates automatically. Defaults to `false`.",
		},
		"trust_fingerprints": resourceSchema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Specifies whether Fivetran should trust SSH fingerprints automatically. Defaults to `false`.",
		},
	}
}
