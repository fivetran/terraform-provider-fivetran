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

func ProxyAgent() resource.Resource {
    return &proxy{}
}

type proxy struct {
    core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &proxy{}
var _ resource.ResourceWithImportState = &proxy{}

func (r *proxy) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_proxy_agent"
}

func (r *proxy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = fivetranSchema.ProxyAgentResource()
}

func (r *proxy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *proxy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ProxyAgentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewProxyCreate()
	svc.DisplayName(data.DisplayName.ValueString())
	svc.GroupRegion(data.GroupRegion.ValueString())

	createResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Proxy Agent Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, createResponse.Code, createResponse.Message),
		)

		return
	}

	data.ReadFromCreateResponse(createResponse)

    readResponse, err := r.GetClient().NewProxyDetails().ProxyId(data.Id.ValueString()).Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Proxy Agent Resource.",
            fmt.Sprintf("%v; code: %v", err, readResponse.Code),
        )
        return
    }

    data.ReadFromResponse(readResponse)
    
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *proxy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.ProxyAgentResourceModel

    // Read Terraform prior state data into the model
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    readResponse, err := r.GetClient().NewProxyDetails().ProxyId(data.Id.ValueString()).Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Proxy Agent Resource.",
            fmt.Sprintf("%v; code: %v", err, readResponse.Code),
        )
        return
    }

    data.ReadFromResponse(readResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *proxy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    resp.Diagnostics.AddError(
        "Modification does not support",
        "Modification does not support",
    )

    return
}

func (r *proxy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    if r.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.ProxyAgentResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

    deleteResponse, err := r.GetClient().NewProxyDelete().ProxyId(data.Id.ValueString()).Do(ctx)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Delete Proxy Agent Resource.",
            fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
        )
        return
    }
}