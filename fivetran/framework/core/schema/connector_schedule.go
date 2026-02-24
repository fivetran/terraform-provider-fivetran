package schema

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func GetConnectorScheduleResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique resource identifier (equals to `connector_id`).",
			},
			"connector_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier for the connector within the Fivetran system.",
			},
			"group_id": schema.StringAttribute{
				Optional:    true,
				Description: "The unique identifier for the Group (Destination) within the Fivetran system.",
			},
			"connector_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name used both as the connection's name within the Fivetran system and as the source schema's name within your destination.",
			},
			"sync_frequency": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("1", "5", "15", "30", "60", "120", "180", "360", "480", "720", "1440"),
				},
				Description: "The connector sync frequency in minutes. Supported values: 1, 5, 15, 30, 60, 120, 180, 360, 480, 720, 1440. Deprecated: use `schedule` block instead.",
			},
			"schedule_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The connector schedule configuration type. Supported values: auto, manual.",
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "manual"),
				},
			},
			"paused": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Specifies whether the connector is paused.",
				Validators: []validator.String{
					stringvalidator.OneOf("true", "false"),
				},
			},
			"pause_after_trial": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Specifies whether the connector should be paused after the free trial period has ended.",
				Validators: []validator.String{
					stringvalidator.OneOf("true", "false"),
				},
			},
			"daily_sync_time": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use [the baseline sync start time](https://fivetran.com/docs/getting-started/syncoverview#syncfrequencyandscheduling). This parameter has no effect on the [0 to 60 minutes offset](https://fivetran.com/docs/getting-started/syncoverview#syncstarttimesandoffsets) used to determine the actual sync start time.",
				Validators: []validator.String{
					stringvalidator.OneOf("00:00", "01:00", "02:00", "03:00", "04:00", "05:00", "06:00", "07:00", "08:00", "09:00",
						"10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00", "20:00", "21:00", "22:00", "23:00"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"schedule": schema.SingleNestedBlock{
				Description: "Flexible sync schedule configuration. When set, takes precedence over `sync_frequency`.",
				Attributes: map[string]schema.Attribute{
					"schedule_type": schema.StringAttribute{
						Optional:    true,
						Description: "The schedule type. Supported values: INTERVAL, TIME_OF_DAY, CRON, MANUAL.",
						Validators: []validator.String{
							stringvalidator.OneOf("INTERVAL", "TIME_OF_DAY", "CRON", "MANUAL"),
						},
					},
					"interval": schema.Int64Attribute{
						Optional:    true,
						Description: "The sync interval in minutes. Required for INTERVAL schedule type.",
					},
					"time_of_day": schema.StringAttribute{
						Optional:    true,
						Description: `The time of day to run the sync. Required for TIME_OF_DAY schedule type. Supported values: "00:00" to "23:00".`,
					},
					"days_of_week": schema.SetAttribute{
						Optional:    true,
						ElementType: basetypes.StringType{},
						Description: "The days of the week to run the sync. Used with INTERVAL and TIME_OF_DAY schedule types. Supported values: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY.",
					},
					"cron": schema.StringAttribute{
						Optional:    true,
						Description: "The cron expression for the sync schedule. Required for CRON schedule type.",
					},
				},
			},
		},
	}
}
