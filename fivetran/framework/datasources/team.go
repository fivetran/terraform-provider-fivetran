package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func Team() datasource.DataSource {
	return &team{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &team{}

type team struct {
	core.ProviderDatasource
}

func (d *team) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_team"
}

func (d *team) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.TeamDatasource()
}

func (d *team) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Team

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	listResponse, err := d.GetClient().NewTeamsDetails().TeamId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v", err, listResponse.Code),
		)
		return
	}

	data.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}