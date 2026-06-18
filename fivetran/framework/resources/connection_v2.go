package resources

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ConnectionV2() resource.Resource {
	return &connectionV2{}
}

type connectionV2 struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &connectionV2{}

func (r *connectionV2) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_v2"
}

func (r *connectionV2) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.ConnectionV2ResourceSchema()
}

func (r *connectionV2) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"fivetran_connection_v2 is not registered",
		"The resource skeleton is present for development, but CRUD is implemented in a follow-up change and the resource is not registered with the provider yet.",
	)
}

func (r *connectionV2) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError(
		"fivetran_connection_v2 is not registered",
		"The resource skeleton is present for development, but CRUD is implemented in a follow-up change and the resource is not registered with the provider yet.",
	)
}

func (r *connectionV2) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"fivetran_connection_v2 is not registered",
		"The resource skeleton is present for development, but CRUD is implemented in a follow-up change and the resource is not registered with the provider yet.",
	)
}

func (r *connectionV2) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"fivetran_connection_v2 is not registered",
		"The resource skeleton is present for development, but CRUD is implemented in a follow-up change and the resource is not registered with the provider yet.",
	)
}
