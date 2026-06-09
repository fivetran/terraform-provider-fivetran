package resources

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ConnectionV2() resource.Resource {
	return &connectionV2{}
}

type connectionV2 struct {
	core.ProviderResource
}

var _ resource.ResourceWithConfigure = &connectionV2{}
var _ resource.ResourceWithImportState = &connectionV2{}

func (r *connectionV2) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection_v2"
}

func (r *connectionV2) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = connectionV2Schema()
}

func (r *connectionV2) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *connectionV2) Create(_ context.Context, _ resource.CreateRequest, _ *resource.CreateResponse) {}
func (r *connectionV2) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse)       {}
func (r *connectionV2) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {}
func (r *connectionV2) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}
