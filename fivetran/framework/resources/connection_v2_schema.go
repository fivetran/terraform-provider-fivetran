package resources

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func connectionV2Schema() schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			// --- ForceNew ---
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the connection within the Fivetran system.",
			},
			"service": schema.StringAttribute{
				Required:    true,
				Description: "The connector service type (e.g. `google_sheets`, `postgres`). Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the destination group. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_schema": schema.StringAttribute{
				Required:    true,
				Description: "The destination schema identifier. Format depends on the service: a single segment (e.g. `my_schema`), two dot-separated segments (e.g. `my_schema.my_table`), or a prefix (e.g. `my_prefix`). Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// --- Dynamic config slots ---
			"config": schema.DynamicAttribute{
				Optional:    true,
				Description: "The connector configuration. Fields depend on the service; see the Fivetran connector documentation for your service. Validated at plan time against the connector metadata endpoint.",
			},
			"auth": schema.DynamicAttribute{
				Optional:    true,
				Description: "The connector auth configuration. Fields depend on the service.",
			},

			// --- Computed-only ---
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The connector name within the Fivetran system (mirrors the schema portion of destination_schema).",
			},
			"connected_by": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the user who created the connection.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp of when the connection was created.",
			},
			"succeeded_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp of the last successful sync.",
			},
			"failed_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp of the last failed sync.",
			},
			"service_version": schema.StringAttribute{
				Computed:    true,
				Description: "The connector type version within the Fivetran system.",
			},
			"status": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The connector status.",
				Attributes: map[string]schema.Attribute{
					"setup_state": schema.StringAttribute{
						Computed:    true,
						Description: "The current setup state of the connector. Possible values: `incomplete`, `connected`, `broken`.",
					},
					"sync_state": schema.StringAttribute{
						Computed:    true,
						Description: "The current sync state. Possible values: `scheduled`, `syncing`, `paused`, `rescheduled`.",
					},
					"update_state": schema.StringAttribute{
						Computed:    true,
						Description: "The current data update state. Possible values: `on_schedule`, `delayed`.",
					},
					"is_historical_sync": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the next scheduled sync will be a full historical re-sync.",
					},
					"tasks": schema.SetNestedAttribute{
						Computed:    true,
						Description: "Current tasks for the connector.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"code":    schema.StringAttribute{Computed: true},
								"message": schema.StringAttribute{Computed: true},
							},
						},
					},
					"warnings": schema.SetNestedAttribute{
						Computed:    true,
						Description: "Current warnings for the connector.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"code":    schema.StringAttribute{Computed: true},
								"message": schema.StringAttribute{Computed: true},
							},
						},
					},
				},
			},

			// --- Optional + Computed root attributes ---
			"paused": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the connector is paused.",
			},
			"sync_frequency": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The sync frequency in minutes. Supported values: 1, 5, 15, 30, 60, 120, 180, 360, 480, 720, 1440.",
			},
			"schedule_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The schedule type. Supported values: `auto`, `manual`. Note: cannot be set on Create — defaults to `auto` and can be changed on the first Update.",
			},
			"daily_sync_time": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The sync start time when sync_frequency is 1440 (daily). Format: `HH:MM` in one-hour increments from `00:00` to `23:00`.",
			},
			"pause_after_trial": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to pause the connector after the free trial ends.",
			},
			"networking_method": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The networking method. Possible values: `Directly`, `SshTunnel`, `ProxyAgent`, `PrivateLink`.",
			},
			"proxy_agent_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The proxy agent ID. Required when networking_method is `ProxyAgent`.",
			},
			"private_link_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The private link ID. Required when networking_method is `PrivateLink`.",
			},
			"hybrid_deployment_agent_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The hybrid deployment agent ID.",
			},
			"data_delay_sensitivity": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The data delay notification threshold level. Possible values: `LOW`, `NORMAL`, `HIGH`, `CUSTOM`, `SYNC_FREQUENCY`.",
			},
			"data_delay_threshold": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Custom data delay notification threshold in minutes. Only used when data_delay_sensitivity is `CUSTOM`.",
			},

			// --- Plan-only (not round-tripped from API) ---
			"run_setup_tests": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to run setup tests on Create. Not stored in state after creation.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"trust_certificates": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to automatically trust the source certificate.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"trust_fingerprints": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to automatically trust the SSH fingerprint.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

