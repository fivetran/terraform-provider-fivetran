package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)


func TeamGroupMemberships() datasource.DataSource {
	return &teamGroupMemberships{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &teamGroupMemberships{}

type teamGroupMemberships struct {
	core.ProviderDatasource
}

func (d *teamGroupMemberships) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_team_group_memberships"
}

func (d *teamGroupMemberships) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.TeamGroupMembershipDatasource()
}

func (d *teamGroupMemberships) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TeamGroupMemberships
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    teamGroupResponse, err := data.ReadFromSource(ctx, d.GetClient(), data.TeamId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Team Group Memberships DataSource.",
            fmt.Sprintf("%v; code: %v", err, teamGroupResponse.Code),
        )

        return
    }

	data.ReadFromResponse(ctx, teamGroupResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}