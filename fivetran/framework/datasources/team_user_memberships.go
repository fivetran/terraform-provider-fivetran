package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)


func TeamUserMemberships() datasource.DataSource {
	return &teamUserMemberships{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &teamUserMemberships{}

type teamUserMemberships struct {
	core.ProviderDatasource
}

func (d *teamUserMemberships) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_team_user_memberships"
}

func (d *teamUserMemberships) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.TeamUserMembershipDatasource()
}

func (d *teamUserMemberships) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TeamUserMemberships
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    teamUserResponse, err := data.ReadFromSource(ctx, d.GetClient(), data.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Team User Memberships DataSource.",
            fmt.Sprintf("%v; code: %v", err, teamUserResponse.Code),
        )

        return
    }

	data.ReadFromResponse(ctx, teamUserResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}