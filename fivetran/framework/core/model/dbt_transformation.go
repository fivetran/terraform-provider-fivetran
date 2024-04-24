package model

import (
	"context"

	"github.com/fivetran/go-fivetran/dbt"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DbtTransformation struct {
	Id              types.String `tfsdk:"id"`
	DbtProjectId    types.String `tfsdk:"dbt_project_id"`
	DbtModelName    types.String `tfsdk:"dbt_model_name"`
	RunTests        types.Bool   `tfsdk:"run_tests"`
	Paused          types.Bool   `tfsdk:"paused"`
	DbtModelId      types.String `tfsdk:"dbt_model_id"`
	OutputModelName types.String `tfsdk:"output_model_name"`
	CreatedAt       types.String `tfsdk:"created_at"`
	ConnectorIds    types.Set    `tfsdk:"connector_ids"`
	ModelIds        types.Set    `tfsdk:"model_ids"`
	Schedule        types.Object `tfsdk:"schedule"`
}

type DbtTransformationResourceModel struct {
	Id              types.String   `tfsdk:"id"`
	DbtProjectId    types.String   `tfsdk:"dbt_project_id"`
	DbtModelName    types.String   `tfsdk:"dbt_model_name"`
	RunTests        types.Bool     `tfsdk:"run_tests"`
	Paused          types.Bool     `tfsdk:"paused"`
	DbtModelId      types.String   `tfsdk:"dbt_model_id"`
	OutputModelName types.String   `tfsdk:"output_model_name"`
	CreatedAt       types.String   `tfsdk:"created_at"`
	ConnectorIds    types.Set      `tfsdk:"connector_ids"`
	ModelIds        types.Set      `tfsdk:"model_ids"`
	Schedule        types.Object   `tfsdk:"schedule"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

func (d *DbtTransformationResourceModel) ReadFromResponse(ctx context.Context, resp dbt.DbtTransformationResponse, modelResp *dbt.DbtModelDetailsResponse) {
	d.Id = types.StringValue(resp.Data.ID)
	d.DbtProjectId = types.StringValue(resp.Data.DbtProjectId)

	if modelResp != nil {
		d.DbtModelName = types.StringValue(modelResp.Data.ModelName)
	} else {
		if d.DbtModelName.IsNull() || d.DbtModelName.IsUnknown() {
			d.DbtModelName = types.StringNull()
		}
	}

	d.DbtModelId = types.StringValue(resp.Data.DbtModelId)
	d.RunTests = types.BoolValue(resp.Data.RunTests)
	d.Paused = types.BoolValue(resp.Data.Paused)
	d.OutputModelName = types.StringValue(resp.Data.OutputModelName)
	d.CreatedAt = types.StringValue(resp.Data.CreatedAt)

	if resp.Data.ModelIds != nil {
		d.ModelIds = types.SetValueMust(types.StringType, stringListToAttrList(resp.Data.ModelIds))
	} else {
		d.ModelIds = types.SetNull(types.StringType)
	}

	if resp.Data.ConnectorIds != nil {
		d.ConnectorIds = types.SetValueMust(types.StringType, stringListToAttrList(resp.Data.ConnectorIds))
	} else {
		d.ConnectorIds = types.SetNull(types.StringType)
	}

	scheduleAttrs := map[string]attr.Type{
		"schedule_type": types.StringType,
		"days_of_week":  types.SetType{ElemType: types.StringType},
		"interval":      types.Int64Type,
		"time_of_day":   types.StringType,
	}

	var daysOfWeekValue attr.Value

	if resp.Data.Schedule.DaysOfWeek != nil {
		daysOfWeekValue = types.SetValueMust(types.StringType, stringListToAttrList(resp.Data.Schedule.DaysOfWeek))
	} else {
		daysOfWeekValue = types.SetNull(types.StringType)
	}
	scheduleAttrValues := map[string]attr.Value{
		"schedule_type": types.StringValue(resp.Data.Schedule.ScheduleType),
	}
	if resp.Data.Schedule.ScheduleType == "INTERVAL" || resp.Data.Schedule.Interval > 0 {
		scheduleAttrValues["interval"] = types.Int64Value(int64(resp.Data.Schedule.Interval))
	} else {
		scheduleAttrValues["interval"] = types.Int64Null()
	}
	if resp.Data.Schedule.TimeOfDay != "" {
		scheduleAttrValues["time_of_day"] = types.StringValue(resp.Data.Schedule.TimeOfDay)
	} else {
		scheduleAttrValues["time_of_day"] = types.StringNull()
	}
	if len(resp.Data.Schedule.DaysOfWeek) > 0 {
		scheduleAttrValues["days_of_week"] = daysOfWeekValue
	} else {
		scheduleAttrValues["days_of_week"] = types.SetNull(types.StringType)
	}

	d.Schedule = types.ObjectValueMust(scheduleAttrs, scheduleAttrValues)
}

func (d *DbtTransformation) ReadFromResponse(ctx context.Context, resp dbt.DbtTransformationResponse, modelResp *dbt.DbtModelDetailsResponse) {
	d.Id = types.StringValue(resp.Data.ID)
	d.DbtProjectId = types.StringValue(resp.Data.DbtProjectId)

	if modelResp != nil {
		d.DbtModelName = types.StringValue(modelResp.Data.ModelName)
	} else {
		if d.DbtModelName.IsNull() || d.DbtModelName.IsUnknown() {
			d.DbtModelName = types.StringNull()
		}
	}

	d.DbtModelId = types.StringValue(resp.Data.DbtModelId)
	d.RunTests = types.BoolValue(resp.Data.RunTests)
	d.Paused = types.BoolValue(resp.Data.Paused)
	d.OutputModelName = types.StringValue(resp.Data.OutputModelName)
	d.CreatedAt = types.StringValue(resp.Data.CreatedAt)

	if resp.Data.ModelIds != nil {
		d.ModelIds = types.SetValueMust(types.StringType, stringListToAttrList(resp.Data.ModelIds))
	} else {
		d.ModelIds = types.SetNull(types.StringType)
	}

	if resp.Data.ConnectorIds != nil {
		d.ConnectorIds = types.SetValueMust(types.StringType, stringListToAttrList(resp.Data.ConnectorIds))
	} else {
		d.ConnectorIds = types.SetNull(types.StringType)
	}

	scheduleAttrs := map[string]attr.Type{
		"schedule_type": types.StringType,
		"days_of_week":  types.SetType{ElemType: types.StringType},
		"interval":      types.Int64Type,
		"time_of_day":   types.StringType,
	}

	var daysOfWeekValue attr.Value

	if resp.Data.Schedule.DaysOfWeek != nil {
		daysOfWeekValue = types.SetValueMust(types.StringType, stringListToAttrList(resp.Data.Schedule.DaysOfWeek))
	} else {
		daysOfWeekValue = types.SetNull(types.StringType)
	}

	scheduleAttrValues := map[string]attr.Value{
		"schedule_type": types.StringValue(resp.Data.Schedule.ScheduleType),
		"days_of_week":  daysOfWeekValue,
		"interval":      types.Int64Value(int64(resp.Data.Schedule.Interval)),
		"time_of_day":   types.StringValue(resp.Data.Schedule.TimeOfDay),
	}

	d.Schedule = types.ObjectValueMust(scheduleAttrs, scheduleAttrValues)
}

func stringListToAttrList(in []string) []attr.Value {
	result := []attr.Value{}
	for _, i := range in {
		result = append(result, types.StringValue(i))
	}
	return result
}
