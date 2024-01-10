package schema

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func GetConnectorScheduleResourceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique resource identifier (equals to `connector_id`).",
			},
			"connector_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The unique identifier for the connector within the Fivetran system.",
			},
			"sync_frequency": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("1", "5", "15", "30", "60", "120", "180", "360", "480", "720", "1440"),
				},
				Description: "The connector sync frequency in minutes. Supported values: 1, 5, 15, 30, 60, 120, 180, 360, 480, 720, 1440.",
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
	}
}
