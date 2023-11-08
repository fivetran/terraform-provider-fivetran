package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func User() datasource.DataSource {
	return &user{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &user{}

type user struct {
	core.ProviderDatasource
}

func (d *user) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_user"
}

func (d *user) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: fivetranSchema.User().GetDatasourceSchema(),
	}
}

func (d *user) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.User

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	userResponse, err := d.GetClient().NewUserDetails().UserID(data.ID.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, userResponse.Code, userResponse.Message),
		)
		return
	}

	data.ReadFromResponse(userResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
