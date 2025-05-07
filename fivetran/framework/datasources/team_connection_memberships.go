package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)


func TeamConnectionMemberships() datasource.DataSource {
	return &teamConnectionMemberships{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &teamConnectionMemberships{}

type teamConnectionMemberships struct {
	core.ProviderDatasource
}

func (d *teamConnectionMemberships) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_team_connection_memberships"
}

func (d *teamConnectionMemberships) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.TeamConnectionMembershipDatasource()
}

func (d *teamConnectionMemberships) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TeamConnectionMemberships
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    listResponse, err := data.ReadFromSource(ctx, d.GetClient(), data.Id.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Team Connection Memberships DataSource.",
            fmt.Sprintf("%v; code: %v", err, listResponse.Code),
        )

        return
    }

	data.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}