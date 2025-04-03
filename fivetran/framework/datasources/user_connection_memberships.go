package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)


func UserConnectorMemberships() datasource.DataSource {
	return &userConnectorMemberships{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &userConnectorMemberships{}

type userConnectorMemberships struct {
	core.ProviderDatasource
}

func (d *userConnectorMemberships) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_user_connector_memberships"
}

func (d *userConnectorMemberships) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.UserConnectorMembershipDatasource()
}

func (d *userConnectorMemberships) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.UserConnectorMemberships
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    userConnectorResponse, err := data.ReadFromSource(ctx, d.GetClient(), data.UserId.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read User Connector Memberships DataSource.",
            fmt.Sprintf("%v; code: %v", err, userConnectorResponse.Code),
        )

        return
    }

	data.ReadFromResponse(ctx, userConnectorResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}