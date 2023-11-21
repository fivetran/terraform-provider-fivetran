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

func GroupSshKey() datasource.DataSource {
	return &groupSshKey{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &groupSshKey{}

type groupSshKey struct {
	core.ProviderDatasource
}

func (d *groupSshKey) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_group_ssh_key"
}

func (d *groupSshKey) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: fivetranSchema.GroupSshKey().GetDatasourceSchema(),
	}
}

func (d *groupSshKey) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.GroupSshKey

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	groupSshKeyResponse, err := d.GetClient().NewGroupSshPublicKey().GroupID(data.ID.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, groupSshKeyResponse.Code, groupSshKeyResponse.Message),
		)
		return
	}

	data.ReadFromResponse(groupSshKeyResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
