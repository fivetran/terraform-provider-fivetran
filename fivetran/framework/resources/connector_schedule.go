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

func ConnectorSchedule() resource.Resource {
	return &connectorSchedule{}
}

type connectorSchedule struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &connectorSchedule{}

func (r *connectorSchedule) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector_schedule"
}

func (r *connectorSchedule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.GetConnectorScheduleResourceSchema()
}

func (r *connectorSchedule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectorSchedule

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewConnectionUpdate().
		ConnectionID(data.ConnectorId.ValueString())

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

	connectorResponse, err := svc.DoCustom(ctx)

	if err != nil {
		if connectorResponse.Code == "NotFound_Connector" {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schedule Resource.",
				"Connector with id = "+data.ConnectorId.ValueString()+" doesn't exist.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Create Connector Schedule Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, connectorResponse.Code, connectorResponse.Message),
			)
		}
		return
	}

	data.ReadFromUpdateResponse(connectorResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorSchedule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectorSchedule

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	connectorResponse, err := r.GetClient().NewConnectionDetails().ConnectionID(data.Id.ValueString()).DoCustom(ctx)

	if err != nil {
		if connectorResponse.Code == "NotFound_Connector" {
			resp.Diagnostics.AddError(
				"Unable to Read Connector Schedule Resource.",
				"Connector with id = "+data.ConnectorId.ValueString()+" doesn't exist.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Read Connector Schedule Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, connectorResponse.Code, connectorResponse.Message),
			)
		}
		return
	}
	data.ReadFromResponse(connectorResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectorSchedule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.ConnectorSchedule

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	svc := r.GetClient().NewConnectionUpdate().
		ConnectionID(state.ConnectorId.ValueString())

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

	connectorResponse, err := svc.DoCustom(ctx)

	if err != nil {
		if connectorResponse.Code == "NotFound_Connector" {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Schedule Resource.",
				"Connector with id = "+state.ConnectorId.ValueString()+" doesn't exist.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Unable to Update Connector Schedule Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, connectorResponse.Code, connectorResponse.Message),
			)
		}
		return
	}

	state.ReadFromUpdateResponse(connectorResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *connectorSchedule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// do nothing
}
