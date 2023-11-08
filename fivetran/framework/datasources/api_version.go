package datasources

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ApiVersion() datasource.DataSource {
	return &apiVersion{}
}

var _ datasource.DataSourceWithConfigure = &apiVersion{}

type apiVersion struct {
	core.ProviderDatasource
}

func (d *apiVersion) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_api_version"
}

func (d *apiVersion) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *apiVersion) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id types.String

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &id)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("version"), "someVersion")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
