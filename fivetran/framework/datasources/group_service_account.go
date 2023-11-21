package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func GroupServiceAccount() datasource.DataSource {
	return &groupServiceAccount{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &groupServiceAccount{}

type groupServiceAccount struct {
	core.ProviderDatasource
}

func (d *groupServiceAccount) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_group_service_account"
}

func (d *groupServiceAccount) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: fivetranSchema.GroupServiceAccount().GetDatasourceSchema(),
	}
}

func (d *groupServiceAccount) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.GroupServiceAccount

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	groupServiceAccountResponse, err := d.GetClient().NewGroupServiceAccount().GroupID(data.ID.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, groupServiceAccountResponse.Code, groupServiceAccountResponse.Message),
		)
		return
	}

	data.ReadFromResponse(groupServiceAccountResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
