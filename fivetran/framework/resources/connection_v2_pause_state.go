package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran/connections"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ConnectionV2PauseState() resource.Resource {
	return &connectionV2PauseState{}
}

type connectionV2PauseState struct {
	core.ProviderResource
}

const connectionV2PauseStateUserAgentSuffix = "fivetran_connection_v2_pause_state"

var _ resource.ResourceWithConfigure = &connectionV2PauseState{}
var _ resource.ResourceWithImportState = &connectionV2PauseState{}

func (r *connectionV2PauseState) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_v2_pause_state"
}

func (r *connectionV2PauseState) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.ConnectionV2PauseStateResourceSchema()
}

func (r *connectionV2PauseState) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	data, details, err := r.readPauseState(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Import Connection V2 Pause State Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, details.Code, details.Message),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionV2PauseState) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.ConnectionV2PauseStateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	update, err := r.setPauseState(ctx, data.ConnectionId.ValueString(), data.Paused.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection V2 Pause State Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, update.Code, update.Message),
		)
		return
	}

	observed, details, err := r.readPauseState(ctx, data.ConnectionId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Connection V2 Pause State Resource After Create.",
			fmt.Sprintf("%v; code: %v; message: %v", err, details.Code, details.Message),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &observed)...)
}

func (r *connectionV2PauseState) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var data model.ConnectionV2PauseStateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionID := data.ConnectionId.ValueString()
	if connectionID == "" {
		connectionID = data.Id.ValueString()
	}

	observed, details, err := r.readPauseState(ctx, connectionID)
	if err != nil {
		if isConnectionV2PauseStateNotFound(details.Code) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Unable to Read Connection V2 Pause State Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, details.Code, details.Message),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &observed)...)
}

func (r *connectionV2PauseState) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	var plan, state model.ConnectionV2PauseStateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Paused.Equal(state.Paused) {
		update, err := r.setPauseState(ctx, plan.ConnectionId.ValueString(), plan.Paused.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Connection V2 Pause State Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, update.Code, update.Message),
			)
			return
		}
	}

	observed, details, err := r.readPauseState(ctx, plan.ConnectionId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Connection V2 Pause State Resource After Update.",
			fmt.Sprintf("%v; code: %v; message: %v", err, details.Code, details.Message),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &observed)...)
}

func (r *connectionV2PauseState) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Intentionally do not modify the underlying connection when this companion resource is removed.
}

func (r *connectionV2PauseState) setPauseState(ctx context.Context, connectionID string, paused bool) (connections.DetailsWithCustomConfigResponse, error) {
	return r.GetClient().
		NewConnectionUpdate().
		ConnectionID(connectionID).
		Paused(paused).
		DoCustomWithUserAgentSuffix(ctx, connectionV2PauseStateUserAgentSuffix)
}

func (r *connectionV2PauseState) readPauseState(ctx context.Context, connectionID string) (model.ConnectionV2PauseStateResourceModel, connections.DetailsWithCustomConfigNoTestsResponse, error) {
	response, err := r.GetClient().
		NewConnectionDetails().
		ConnectionID(connectionID).
		DoCustomWithUserAgentSuffix(ctx, connectionV2PauseStateUserAgentSuffix)

	var data model.ConnectionV2PauseStateResourceModel
	if err == nil {
		data.ReadFromResponse(response)
	}
	return data, response, err
}

func isConnectionV2PauseStateNotFound(code string) bool {
	return code == "NotFound_Connector" || code == "NotFound_Connection"
}
