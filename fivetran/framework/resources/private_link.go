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

func PrivateLink() resource.Resource {
	return &privateLink{}
}

type privateLink struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &privateLink{}
var _ resource.ResourceWithImportState = &privateLink{}

func (r *privateLink) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_link"
}

func (r *privateLink) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.PrivateLinkResource()
}

func (r *privateLink) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *privateLink) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.PrivateLink

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewPrivateLinkCreate()
	svc.Region(data.Region.ValueString())
	svc.Service(data.Service.ValueString())
	svc.Name(data.Name.ValueString())

	config := data.GetConfig()
	svc.Config(&config)

	createResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Private Link Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, createResponse.Code, createResponse.Message),
		)

		return
	}

	data.ReadFromResponse(ctx, createResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *privateLink) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.PrivateLink

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	readResponse, err := r.GetClient().NewPrivateLinkDetails().PrivateLinkId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Private Link Resource.",
			fmt.Sprintf("%v; code: %v", err, readResponse.Code),
		)
		return
	}

	data.ReadFromResponse(ctx, readResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *privateLink) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Private Link does not support",
		"Please report this issue to the provider developers.",
	)
	return

	var plan, state model.PrivateLink
	hasChanges := false

	svc := r.GetClient().NewPrivateLinkModify().PrivateLinkId(state.Id.ValueString())

	if !plan.Config.Equal(state.Config) {
		config := plan.GetConfig()
		svc.Config(&config)
		hasChanges = true
	}

	if hasChanges {
		updateResponse, err := svc.Do(ctx)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Private Link Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, updateResponse.Code, updateResponse.Message),
			)
			return
		}

		state.ReadFromResponse(ctx, updateResponse)
	}


	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *privateLink) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.PrivateLink

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	deleteResponse, err := r.GetClient().NewPrivateLinkDelete().PrivateLinkId(data.Id.ValueString()).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Private Link Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}
