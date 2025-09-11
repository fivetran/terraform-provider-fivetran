package model

import (
	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ConnectorSchedule struct {
	Id              types.String `tfsdk:"id"`
	ConnectorId     types.String `tfsdk:"connector_id"`
	GroupId         types.String `tfsdk:"group_id"`
	ConnectorName   types.String `tfsdk:"connector_name"`
	SyncFrequency   types.String `tfsdk:"sync_frequency"`
	ScheduleType    types.String `tfsdk:"schedule_type"`
	Paused          types.String `tfsdk:"paused"`
	PauseAfterTrial types.String `tfsdk:"pause_after_trial"`
	DailySyncTime   types.String `tfsdk:"daily_sync_time"`
}

func (d *ConnectorSchedule) ReadFromResponse(response connections.DetailsWithCustomConfigNoTestsResponse) {
	d.Id = types.StringValue(response.Data.ID)
	d.ConnectorId = types.StringValue(response.Data.ID)
	d.SyncFrequency = types.StringValue(helpers.IntPointerToStr(response.Data.SyncFrequency))
	d.PauseAfterTrial = types.StringValue(helpers.BoolPointerToStr(response.Data.PauseAfterTrial))
	d.Paused = types.StringValue(helpers.BoolPointerToStr(response.Data.Paused))
	d.ScheduleType = types.StringValue(response.Data.ScheduleType)

	if response.Data.SyncFrequency != nil && *response.Data.SyncFrequency == 1440 {
		d.DailySyncTime = types.StringValue(response.Data.DailySyncTime)
	} else if d.DailySyncTime.IsUnknown() {
		d.DailySyncTime = types.StringNull()
	}
}

func (d *ConnectorSchedule) ReadFromUpdateResponse(response connections.DetailsWithCustomConfigResponse) {
	d.Id = types.StringValue(response.Data.ID)
	d.ConnectorId = types.StringValue(response.Data.ID)
	d.PauseAfterTrial = types.StringValue(helpers.BoolPointerToStr(response.Data.PauseAfterTrial))
	d.Paused = types.StringValue(helpers.BoolPointerToStr(response.Data.Paused))
	d.SyncFrequency = types.StringValue(helpers.IntPointerToStr(response.Data.SyncFrequency))
	d.ScheduleType = types.StringValue(response.Data.ScheduleType)

	if response.Data.SyncFrequency != nil && *response.Data.SyncFrequency == 1440 {
		d.DailySyncTime = types.StringValue(response.Data.DailySyncTime)
	} else if d.DailySyncTime.IsUnknown() {
		d.DailySyncTime = types.StringNull()
	}
}
