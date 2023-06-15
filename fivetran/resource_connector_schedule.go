package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceConnectorSchedule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectorScheduleCreate,
		ReadContext:   resourceConnectorScheduleRead,
		UpdateContext: resourceConnectorScheduleUpdate,
		DeleteContext: resourceConnectorScheduleDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			ID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier for the user within the account.",
			},
			CONNECTOR_ID: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The unique identifier for the connector",
			},

			"sync_frequency": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The connector sync frequency in minutes",
			}, // Default: 360
			"schedule_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The connector schedule configuration type. Supported values: auto, manual",
			}, // Default: AUTO
			"paused": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Specifies whether the connector is paused",
			}, // Default: false
			"pause_after_trial": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Specifies whether the connector should be paused after the free trial period has ended",
			}, // Default: false
			"daily_sync_time": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The optional parameter that defines the sync start time when the sync frequency is already set or being set by the current request to 1440. It can be specified in one hour increments starting from 00:00 to 23:00. If not specified, we will use [the baseline sync start time](https://fivetran.com/docs/getting-started/syncoverview#syncfrequencyandscheduling). This parameter has no effect on the [0 to 60 minutes offset](https://fivetran.com/docs/getting-started/syncoverview#syncstarttimesandoffsets) used to determine the actual sync start time",
			},
		},
	}
}

func resourceConnectorScheduleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	connectorId := d.Get(CONNECTOR_ID).(string)
	client := m.(*fivetran.Client)

	// Check if connector exists
	resp, err := client.NewConnectorDetails().ConnectorID(connectorId).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "Connector with id ="+connectorId+" doesn't exist.", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	svc := client.NewConnectorModify().ConnectorID(connectorId)
	if d.Get("sync_frequency").(string) != "" {
		svc.SyncFrequency(strToInt(d.Get("sync_frequency").(string)))
	}
	if d.Get("schedule_type").(string) != "" {
		svc.ScheduleType(d.Get("schedule_type").(string))
	}
	if d.Get("paused").(string) != "" {
		svc.Paused(strToBool(d.Get("paused").(string)))
	}
	if d.Get("pause_after_trial").(string) != "" {
		svc.PauseAfterTrial(strToBool(d.Get("pause_after_trial").(string)))
	}

	if d.Get("sync_frequency") == "1440" && d.Get("daily_sync_time").(string) != "" {
		svc.DailySyncTime(d.Get("daily_sync_time").(string))
	}

	mResp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, mResp.Code, mResp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceConnectorScheduleRead(ctx, d, m)

	return diags
}
func resourceConnectorScheduleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	connectorId := d.Get(CONNECTOR_ID).(string)
	client := m.(*fivetran.Client)

	// Fetch connector
	resp, err := client.NewConnectorDetails().ConnectorID(connectorId).Do(ctx)
	if err != nil {
		if resp.Code == "404" {
			d.SetId("")
			return nil
		}
		return newDiagAppend(diags, diag.Error, "Connector with id ="+connectorId+" doesn't exist.", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	msi := make(map[string]interface{})

	mapAddStr(msi, "sync_frequency", intPointerToStr(resp.Data.SyncFrequency))
	mapAddStr(msi, "schedule_type", resp.Data.ScheduleType)
	mapAddStr(msi, "paused", boolPointerToStr(resp.Data.Paused))
	mapAddStr(msi, "pause_after_trial", boolPointerToStr(resp.Data.PauseAfterTrial))

	// Value for daily_sync_time won't be returned if sync_frequency < 1440 so we can get it from current config to avoid drifting change
	if *resp.Data.SyncFrequency != 1440 {
		mapAddStr(msi, "daily_sync_time", d.Get("daily_sync_time").(string))
	} else {
		mapAddStr(msi, "daily_sync_time", resp.Data.DailySyncTime)
	}

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)
	return diags
}
func resourceConnectorScheduleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewConnectorModify()

	svc.ConnectorID(d.Get("id").(string))

	if d.HasChange("sync_frequency") {
		svc.SyncFrequency(strToInt(d.Get("sync_frequency").(string)))
	}
	if d.HasChange("schedule_type") {
		svc.ScheduleType(d.Get("schedule_type").(string))
	}
	if d.HasChange("paused") {
		svc.Paused(strToBool(d.Get("paused").(string)))
	}
	if d.HasChange("pause_after_trial") {
		svc.PauseAfterTrial(strToBool(d.Get("pause_after_trial").(string)))
	}
	if d.Get("sync_frequency") == "1440" && d.HasChange("daily_sync_time") {
		svc.DailySyncTime(d.Get("daily_sync_time").(string))
	}

	resp, err := svc.Do(ctx)

	if err != nil {
		// resourceConnectorScheduleRead here makes sure the state is updated after a NewConnectorModify error.
		diags = resourceConnectorScheduleRead(ctx, d, m)
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	return resourceConnectorScheduleRead(ctx, d, m)
}
func resourceConnectorScheduleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// nothing to delete
	return diags
}
