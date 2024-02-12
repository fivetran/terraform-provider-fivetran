package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)


func TeamConnectorMemberships() datasource.DataSource {
	return &teamConnectorMemberships{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &teamConnectorMemberships{}

type teamConnectorMemberships struct {
	core.ProviderDatasource
}

func (d *teamConnectorMemberships) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_team_connector_memberships"
}

func (d *teamConnectorMemberships) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.TeamConnectorMembershipDatasource()
}

func (d *teamConnectorMemberships) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TeamConnectorMemberships
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    teamConnectorResponse, err := data.ReadFromSource(ctx, d.GetClient(), data.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Team Connector Memberships DataSource.",
            fmt.Sprintf("%v; code: %v", err, teamConnectorResponse.Code),
        )

        return
    }

	data.ReadFromResponse(ctx, teamConnectorResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}