package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func Group() datasource.DataSource {
	return &group{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &group{}

type group struct {
	core.ProviderDatasource
}

func (d *group) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_group"
}

func (d *group) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.GroupDatasource()
}

func (d *group) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Group

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	groupResponse, err := d.GetClient().NewGroupDetails().GroupID(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, groupResponse.Code, groupResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, groupResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
