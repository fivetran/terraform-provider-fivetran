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

func Team() resource.Resource {
    return &team{}
}

type team struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &team{}
var _ resource.ResourceWithImportState = &team{}

func (r *team) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *team) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.TeamResource()
}

func (r *team) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *team) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Team

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewTeamsCreate()
	svc.Name(data.Name.ValueString())
	svc.Role(data.Role.ValueString())

	if !data.Description.IsUnknown() && !data.Description.IsNull() {
		svc.Description(data.Description.ValueString())
	}

	createResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Team Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, createResponse.Code, createResponse.Message),
		)

		return
	}

	data.ReadFromCreateResponse(ctx, createResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *team) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.Team

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    readResponse, err := r.GetClient().NewTeamsDetails().TeamId(data.Id.ValueString()).Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Team Resource.",
            fmt.Sprintf("%v; code: %v", err, readResponse.Code),
        )
        return
    }

    data.ReadFromResponse(ctx, readResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *team) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.Team

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    svc := r.GetClient().NewTeamsModify().TeamId(state.Id.ValueString())
    
    if !plan.Name.Equal(state.Name) {
        svc.Name(plan.Name.ValueString())
    }

    if !plan.Description.Equal(state.Description) {
        svc.Description(plan.Description.ValueString())
    }

    if !plan.Role.Equal(state.Role) {
        svc.Role(plan.Role.ValueString())
    }

    updateResponse, err := svc.Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Update Team Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
        )
        return
    }

    state.ReadFromModifyResponse(ctx, updateResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *team) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.Team

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    deleteResponse, err := r.GetClient().NewTeamsDelete().TeamId(data.Id.ValueString()).Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Delete Team Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
        )
        return
    }
}