package model

import (
	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ConnectorScheduleBlock represents the nested schedule block in fivetran_connector_schedule.
type ConnectorScheduleBlock struct {
	ScheduleType types.String `tfsdk:"schedule_type"`
	Interval     types.Int64  `tfsdk:"interval"`
	TimeOfDay    types.String `tfsdk:"time_of_day"`
	DaysOfWeek   types.Set    `tfsdk:"days_of_week"`
	Cron         types.String `tfsdk:"cron"`
}

var connectorScheduleBlockAttrTypes = map[string]attr.Type{
	"schedule_type": types.StringType,
	"interval":      types.Int64Type,
	"time_of_day":   types.StringType,
	"days_of_week":  types.SetType{ElemType: types.StringType},
	"cron":          types.StringType,
}

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
	Schedule        types.Object `tfsdk:"schedule"`
}

func readScheduleFromResponse(s *connections.ConnectorSchedule, existing types.Object) types.Object {
	if s == nil {
		return types.ObjectNull(connectorScheduleBlockAttrTypes)
	}

	vals := map[string]attr.Value{}

	if s.ScheduleType != nil {
		vals["schedule_type"] = types.StringValue(*s.ScheduleType)
	} else {
		vals["schedule_type"] = types.StringNull()
	}

	if s.Interval != nil {
		vals["interval"] = types.Int64Value(int64(*s.Interval))
	} else {
		vals["interval"] = types.Int64Null()
	}

	if s.TimeOfDay != nil {
		vals["time_of_day"] = types.StringValue(*s.TimeOfDay)
	} else {
		vals["time_of_day"] = types.StringNull()
	}

	if s.Cron != nil {
		vals["cron"] = types.StringValue(*s.Cron)
	} else {
		vals["cron"] = types.StringNull()
	}

	if len(s.DaysOfWeek) > 0 {
		elems := make([]attr.Value, len(s.DaysOfWeek))
		for i, d := range s.DaysOfWeek {
			elems[i] = types.StringValue(d)
		}
		vals["days_of_week"] = types.SetValueMust(types.StringType, elems)
	} else {
		vals["days_of_week"] = types.SetNull(types.StringType)
	}

	return types.ObjectValueMust(connectorScheduleBlockAttrTypes, vals)
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

	d.Schedule = readScheduleFromResponse(response.Data.Schedule, d.Schedule)
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

	d.Schedule = readScheduleFromResponse(response.Data.Schedule, d.Schedule)
}
