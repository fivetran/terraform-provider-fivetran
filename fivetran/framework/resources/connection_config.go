package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ConnectionConfig() resource.Resource {
	return &connectionConfig{}
}

type connectionConfig struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &connection{}
var _ resource.ResourceWithImportState = &connection{}

func (r *connectionConfig) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_config"
}

func (r *connectionConfig) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.ConnectionConfigResourceSchema()
	resp.Schema.Version = 1
}

func (r *connectionConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("connection_id"), req, resp)
}

func (r *connectionConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectionConfigModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	configMap, authMap, err := data.Validate(ctx, r.GetClient())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Config Resource.",
			err.Error(),
		)

		return
	}

	svc := r.GetClient().NewConnectionUpdate().
		ConnectionID(data.ConnectionId.ValueString()).
		RunSetupTests(true)

	if authMap != nil {
		svc.AuthCustom(&authMap)
	}

	if configMap != nil {
		svc.ConfigCustom(&configMap)
	}

	response, err := svc.
		DoCustom(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Connection Config Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)

		return
	}

	data.Id = data.ConnectionId

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ConnectionConfigModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	response, err := r.GetClient().
		NewConnectionDetails().
		ConnectionID(data.ConnectionId.ValueString()).
		Do(ctx)

	if err != nil {
		// If connection not found, remove from state
		if response.Code == "NotFound_Connector" || response.Code == "NotFound_Connection" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v", err, response.Code),
		)
		return
	}

	data.Id = data.ConnectionId

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.ConnectionConfigModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if plan.Config.Equal(state.Config) && plan.Auth.Equal(state.Auth) {
		return
	}

	configMap, authMap, err := plan.Validate(ctx, r.GetClient())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connection Config Resource.",
			err.Error(),
		)

		return
	}

	svc := r.GetClient().NewConnectionUpdate().
		ConnectionID(state.ConnectionId.ValueString()).
		RunSetupTests(true)

	if authMap != nil && !plan.Auth.Equal(state.Auth) {
		svc.AuthCustom(&authMap)
	}

	if configMap != nil && !plan.Config.Equal(state.Config) {
		svc.ConfigCustom(&configMap)
	}

	response, err := svc.DoCustom(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Connection Config Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, response.Code, response.Message),
		)
		return
	}

	plan.Id = plan.ConnectionId

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *connectionConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// no op
	return
}
