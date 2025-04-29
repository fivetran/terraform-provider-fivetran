package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func User() resource.Resource {
	return &user{}
}

type user struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &user{}
var _ resource.ResourceWithImportState = &user{}

func (r *user) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *user) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: fivetranSchema.User().GetResourceSchema(),
	}
}

func (r *user) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *user) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.User

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewUserInvite()
	svc.Email(data.Email.ValueString())
	svc.GivenName(data.GivenName.ValueString())
	svc.FamilyName(data.FamilyName.ValueString())

	if !data.Phone.IsUnknown() && !data.Phone.IsNull() {
		svc.Phone(data.Phone.ValueString())
	}

	if !data.Role.IsUnknown() && !data.Role.IsNull() {
		svc.Role(data.Role.ValueString())
	}

	if !data.Picture.IsUnknown() && !data.Picture.IsNull() {
		svc.Picture(data.Picture.ValueString())
	}

	userResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create User Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, userResponse.Code, userResponse.Message),
		)

		return
	}

	data.ReadFromResponse(userResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *user) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.User

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	userResponse, err := r.GetClient().NewUserDetails().UserID(data.ID.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read User Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, userResponse.Code, userResponse.Message),
		)
		return
	}

	data.ReadFromResponse(userResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *user) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var plan, state model.User

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	svc := r.GetClient().NewUserUpdate().UserID(state.ID.ValueString())

	if !plan.FamilyName.Equal(state.FamilyName) {
		svc.FamilyName(plan.FamilyName.ValueString())
	}

	if !plan.GivenName.Equal(state.GivenName) {
		svc.GivenName(plan.GivenName.ValueString())
	}

	if !plan.Role.Equal(state.Role) {
		svc.Role(plan.Role.ValueString())
	}
	if !plan.Picture.Equal(state.Picture) {
		if plan.Picture.IsNull() || plan.Picture.IsUnknown() {
			svc.ClearPicture()
		} else {
			svc.Picture(plan.Picture.ValueString())
		}
	}

	if !plan.Phone.Equal(state.Phone) {
		if plan.Phone.IsNull() || plan.Phone.IsUnknown() {
			svc.ClearPhone()
		} else {
			svc.Phone(plan.Phone.ValueString())
		}
	}

	userResponse, err := svc.Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update User Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, userResponse.Code, userResponse.Message),
		)
		return
	}

	state.ReadFromResponse(userResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *user) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.User

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	deleteResponse, err := r.GetClient().NewUserDelete().UserID(data.ID.ValueString()).Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete User Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}
