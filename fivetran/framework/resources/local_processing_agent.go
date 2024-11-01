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

func LocalProcessingAgent() resource.Resource {
    return &localProcessingAgent{}
}

type localProcessingAgent struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &localProcessingAgent{}
var _ resource.ResourceWithImportState = &localProcessingAgent{}

func (r *localProcessingAgent) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_local_processing_agent"
}

func (r *localProcessingAgent) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.LocalProcessingAgentResource()
}

func (r *localProcessingAgent) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *localProcessingAgent) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.LocalProcessingAgentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewLocalProcessingAgentCreate()
	svc.GroupId(data.GroupId.ValueString())
	svc.DisplayName(data.DisplayName.ValueString())
    svc.EnvType("DOCKER")
    svc.AcceptTerms(true)

	createResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Local Processing Agent Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, createResponse.Code, createResponse.Message),
		)

		return
	}

	data.ReadFromCreateResponse(createResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *localProcessingAgent) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.LocalProcessingAgentResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    readResponse, err := r.GetClient().NewLocalProcessingAgentDetails().AgentId(data.Id.ValueString()).Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Local Processing Agent Resource.",
            fmt.Sprintf("%v; code: %v", err, readResponse.Code),
        )
        return
    }

    data.ReadFromResponse(readResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *localProcessingAgent) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.LocalProcessingAgentResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    svc := r.GetClient().NewLocalProcessingAgentReAuth().AgentId(state.Id.ValueString())
    
    updateResponse, err := svc.Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Update Local Processing Agent Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
        )
        return
    }

    state.ReadFromCreateResponse(updateResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *localProcessingAgent) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.LocalProcessingAgentResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    deleteResponse, err := r.GetClient().NewLocalProcessingAgentDelete().AgentId(data.Id.ValueString()).Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Delete Local Processing Agent Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
        )
        return
    }
}