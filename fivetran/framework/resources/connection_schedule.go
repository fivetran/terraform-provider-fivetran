package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ConnectionSchedule() resource.Resource {
	return &connectionSchedule{}
}

type connectionSchedule struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &connectionSchedule{}

func (r *connectionSchedule) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_schedule"
}

func (r *connectionSchedule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.GetConnectionScheduleResourceSchema()
}

func (r *connectionSchedule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectionSchedule

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewConnectionUpdate().
		ConnectionID(data.ConnectionId.ValueString())

	if !data.SyncFrequency.IsNull() && !data.SyncFrequency.IsUnknown() && data.SyncFrequency.ValueString() != "" {
		syncFrequency := helpers.StrToInt(data.SyncFrequency.ValueString())
		svc.SyncFrequency(&syncFrequency)
	} else {
		svc.SyncFrequency(nil)
	}

	if !data.DailySyncTime.IsUnknown() && !data.DailySyncTime.IsNull() && data.SyncFrequency.ValueString() == "1440" {
		svc.DailySyncTime(data.DailySyncTime.ValueString())
	}

	if !data.Paused.IsUnknown() && !data.Paused.IsNull() {
		svc.Paused(helpers.StrToBool(data.Paused.ValueString()))
	}

	if !data.PauseAfterTrial.IsUnknown() && !data.PauseAfterTrial.IsNull() {
		svc.PauseAfterTrial(helpers.StrToBool(data.PauseAfterTrial.ValueString()))
	}

	if !data.ScheduleType.IsUnknown() && !data.ScheduleType.IsNull() {
		svc.ScheduleType(data.ScheduleType.ValueString())
	}

	connectionResponse, err := svc.DoCustom(ctx)

	if err != nil {
		if connectionResponse.Code == "NotFound_Connection" {
			resp.Diagnostics.AddError(
				"Unable to Create Connection Schedule Resource.",
				"Connection with id = "+data.ConnectionId.ValueString()+" doesn't exist.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Create Connection Schedule Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, connectionResponse.Code, connectionResponse.Message),
			)
		}
		return
	}

	data.ReadFromUpdateResponse(connectionResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchedule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectionSchedule

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	connectionResponse, err := r.GetClient().NewConnectionDetails().ConnectionID(data.Id.ValueString()).DoCustom(ctx)

	if err != nil {
		if connectionResponse.Code == "NotFound_Connection" {
			resp.Diagnostics.AddError(
				"Unable to Read Connection Schedule Resource.",
				"Connection with id = "+data.ConnectionId.ValueString()+" doesn't exist.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Read Connection Schedule Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, connectionResponse.Code, connectionResponse.Message),
			)
		}
		return
	}
	data.ReadFromResponse(connectionResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionSchedule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.ConnectionSchedule

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	svc := r.GetClient().NewConnectionUpdate().
		ConnectionID(state.ConnectionId.ValueString())

	if !plan.SyncFrequency.Equal(state.SyncFrequency) {
		if !plan.SyncFrequency.IsNull() && !plan.SyncFrequency.IsUnknown() && plan.SyncFrequency.ValueString() != "" {
			syncFrequency := helpers.StrToInt(plan.SyncFrequency.ValueString())
			svc.SyncFrequency(&syncFrequency)
		} else {
			svc.SyncFrequency(nil)
		}
		if !plan.DailySyncTime.IsUnknown() && plan.SyncFrequency.ValueString() == "1440" {
			svc.DailySyncTime(plan.DailySyncTime.ValueString())
		}
	}

	if !plan.DailySyncTime.IsUnknown() && !plan.DailySyncTime.Equal(state.DailySyncTime) && plan.SyncFrequency.ValueString() == "1440" {
		svc.DailySyncTime(plan.DailySyncTime.ValueString())
	}

	if !plan.Paused.IsUnknown() && !plan.Paused.IsNull() && !plan.Paused.Equal(state.Paused) {
		svc.Paused(helpers.StrToBool(plan.Paused.ValueString()))
	}

	if !plan.PauseAfterTrial.IsUnknown() && !plan.PauseAfterTrial.IsNull() && !plan.PauseAfterTrial.Equal(state.PauseAfterTrial) {
		svc.PauseAfterTrial(helpers.StrToBool(plan.PauseAfterTrial.ValueString()))
	}

	if !plan.ScheduleType.IsUnknown() && !plan.ScheduleType.IsNull() && !plan.ScheduleType.Equal(state.ScheduleType) {
		svc.ScheduleType(plan.ScheduleType.ValueString())
	}

	connectionResponse, err := svc.DoCustom(ctx)

	if err != nil {
		if connectionResponse.Code == "NotFound_Connection" {
			resp.Diagnostics.AddError(
				"Unable to Update Connection Schedule Resource.",
				"Connection with id = "+state.ConnectionId.ValueString()+" doesn't exist.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Update Connection Schedule Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, connectionResponse.Code, connectionResponse.Message),
			)
		}
		return
	}

	state.ReadFromUpdateResponse(connectionResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *connectionSchedule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// do nothing
}
