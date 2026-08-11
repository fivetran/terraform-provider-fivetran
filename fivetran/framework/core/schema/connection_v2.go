package schema

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ConnectionV2ResourceSchema() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: ConnectionV2ResourceAttributes(),
		Version:    0,
	}
}

func ConnectionV2ResourceAttributes() map[string]resourceSchema.Attribute {
	attributes := map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			Computed:    true,
			Description: "The unique identifier for the connection within the Fivetran system.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resourceSchema.StringAttribute{
			Computed:    true,
			Description: "The name used both as the connection's name within the Fivetran system and as the source schema's name within your destination.",
		},
		"connected_by": resourceSchema.StringAttribute{
			Computed:    true,
			Description: "The unique identifier of the user who created the connection in your account.",
		},
		"created_at": resourceSchema.StringAttribute{
			Computed:    true,
			Description: "The timestamp of the time the connection was created in your account.",
		},
		"group_id": resourceSchema.StringAttribute{
			Required:    true,
			Description: "The unique identifier for the Group (Destination) within the Fivetran system.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"service": resourceSchema.StringAttribute{
			Required:    true,
			Description: "The connection service type (e.g., `postgres`, `mysql`, `s3`, `snowflake`). See Fivetran connection types documentation for available services.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"config": resourceSchema.DynamicAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Service-specific connection configuration. The accepted fields are defined by connector metadata at runtime.",
		},
		"auth": resourceSchema.DynamicAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "Service-specific authorization configuration. The accepted fields are defined by connector metadata at runtime.",
		},
		"succeeded_at": resourceSchema.StringAttribute{
			Computed:    true,
			Description: "The timestamp of the time the connection sync succeeded last time.",
		},
		"failed_at": resourceSchema.StringAttribute{
			Computed:    true,
			Description: "The timestamp of the time the connection sync failed last time.",
		},
		"service_version": resourceSchema.StringAttribute{
			Computed:    true,
			Description: "The connection type version within the Fivetran system.",
		},
		"sync_frequency": resourceSchema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: "The connection sync frequency in minutes.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"schedule_type": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The connection schedule configuration type. Supported values: auto, manual.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"pause_after_trial": resourceSchema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Specifies whether the connection should be paused after the free trial period has ended.",
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		"daily_sync_time": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"connect_card_config": resourceSchema.SingleNestedAttribute{
			Optional:    true,
			Description: "Configuration for the interactive Connect Card setup flow.",
			Attributes: map[string]resourceSchema.Attribute{
				"redirect_uri": resourceSchema.StringAttribute{
					Optional:    true,
					Sensitive:   true,
					Description: "URI where the Connect Card flow redirects after setup.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"hide_setup_guide": resourceSchema.BoolAttribute{
					Optional:    true,
					Description: "Whether to hide the setup guide in the Connect Card flow.",
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
				"all_fields": resourceSchema.BoolAttribute{
					Optional:    true,
					Description: "Whether Connect Card should show all setup fields.",
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
					},
				},
			},
		},
		"destination_schema_names": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Defines how schema names appear in the destination. Supported values: FIVETRAN_NAMING, SOURCE_NAMING.",
			Validators: []validator.String{
				stringvalidator.OneOf("FIVETRAN_NAMING", "SOURCE_NAMING"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"destination_configuration": resourceSchema.SingleNestedAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Destination-specific configuration for the connection.",
			Attributes: map[string]resourceSchema.Attribute{
				"virtual_warehouse": resourceSchema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Destination virtual warehouse used by the connection.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
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
		"networking_method": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The networking method for the connection. Possible values: `Directly`, `SshTunnel`, `ProxyAgent`, `PrivateLink`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"hybrid_deployment_agent_id": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The hybrid deployment agent ID that refers to the controller created for the group the connection belongs to.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"private_link_id": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The private link ID. Required when `networking_method` is `PrivateLink`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"data_delay_sensitivity": resourceSchema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The level of data delay notification threshold. Possible values: LOW, NORMAL, HIGH, CUSTOM, SYNC_FREQUENCY.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"data_delay_threshold": resourceSchema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: "Custom sync delay notification threshold in minutes. This parameter is only used when data_delay_sensitivity is set to CUSTOM.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"run_setup_tests": resourceSchema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether to run setup tests when creating or updating the connection.",
		},
		"trust_certificates": resourceSchema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Specifies whether Fivetran should trust certificates automatically.",
		},
		"trust_fingerprints": resourceSchema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Specifies whether Fivetran should trust SSH fingerprints automatically.",
		},
		"status": connectionV2StatusAttribute(),
	}

	return attributes
}

func connectionV2StatusAttribute() resourceSchema.SingleNestedAttribute {
	return resourceSchema.SingleNestedAttribute{
		Computed:    true,
		Description: "The current connection status.",
		Attributes: map[string]resourceSchema.Attribute{
			"setup_state": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The current setup state of the connection.",
			},
			"is_historical_sync": resourceSchema.BoolAttribute{
				Computed:    true,
				Description: "Whether the connection should be triggered to re-sync all historical data on the next scheduled sync.",
			},
			"sync_state": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The current sync state of the connection.",
			},
			"update_state": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The current data update state of the connection.",
			},
			"tasks":    connectionV2CodeMessageSetAttribute("The collection of tasks for the connection."),
			"warnings": connectionV2CodeMessageSetAttribute("The collection of warnings for the connection."),
		},
	}
}

func connectionV2CodeMessageSetAttribute(description string) resourceSchema.SetNestedAttribute {
	return resourceSchema.SetNestedAttribute{
		Computed:    true,
		Description: description,
		NestedObject: resourceSchema.NestedAttributeObject{
			Attributes: map[string]resourceSchema.Attribute{
				"code": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "Code.",
				},
				"message": resourceSchema.StringAttribute{
					Computed:    true,
					Description: "Message.",
				},
			},
		},
	}
}
