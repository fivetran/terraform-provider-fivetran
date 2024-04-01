package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)


func UserGroupMemberships() datasource.DataSource {
	return &userGroupMemberships{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &userGroupMemberships{}

type userGroupMemberships struct {
	core.ProviderDatasource
}

func (d *userGroupMemberships) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_user_group_memberships"
}

func (d *userGroupMemberships) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.UserGroupMembershipDatasource()
}

func (d *userGroupMemberships) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.UserGroupMemberships
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    userGroupResponse, err := data.ReadFromSource(ctx, d.GetClient(), data.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read User Group Memberships DataSource.",
            fmt.Sprintf("%v; code: %v", err, userGroupResponse.Code),
        )

        return
    }

	data.ReadFromResponse(ctx, userGroupResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}