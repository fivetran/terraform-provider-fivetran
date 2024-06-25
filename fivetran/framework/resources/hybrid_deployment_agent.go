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

func HybridDeploymentAgent() resource.Resource {
    return &hybridDeploymentAgent{}
}

type hybridDeploymentAgent struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &hybridDeploymentAgent{}
var _ resource.ResourceWithImportState = &hybridDeploymentAgent{}

func (r *hybridDeploymentAgent) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_hybrid_deployment_agent"
}

func (r *hybridDeploymentAgent) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.HybridDeploymentAgentResource()
}

func (r *hybridDeploymentAgent) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *hybridDeploymentAgent) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.HybridDeploymentAgentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewHybridDeploymentAgentCreate()
	svc.GroupId(data.GroupId.ValueString())
	svc.DisplayName(data.DisplayName.ValueString())
    svc.EnvType("DOCKER")
    svc.AcceptTerms(true)

	createResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Hybrid Deployment Agent Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, createResponse.Code, createResponse.Message),
		)

		return
	}

	data.ReadFromCreateResponse(createResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hybridDeploymentAgent) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.HybridDeploymentAgentResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    readResponse, err := r.GetClient().NewHybridDeploymentAgentDetails().AgentId(data.Id.ValueString()).Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Hybrid Deployment Agent Resource.",
            fmt.Sprintf("%v; code: %v", err, readResponse.Code),
        )
        return
    }

    data.ReadFromResponse(readResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hybridDeploymentAgent) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var plan, state model.HybridDeploymentAgentResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    svc := r.GetClient().NewHybridDeploymentAgentReAuth().AgentId(state.Id.ValueString())
    
    updateResponse, err := svc.Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Update Hybrid Deployment Agent Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
        )
        return
    }

    state.ReadFromCreateResponse(updateResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hybridDeploymentAgent) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.HybridDeploymentAgentResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    deleteResponse, err := r.GetClient().NewHybridDeploymentAgentDelete().AgentId(data.Id.ValueString()).Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Delete Hybrid Deployment Agent Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
        )
        return
    }
}