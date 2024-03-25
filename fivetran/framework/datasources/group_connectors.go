package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func GroupConnectors() datasource.DataSource {
	return &groupConnectors{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &groupConnectors{}

type groupConnectors struct {
	core.ProviderDatasource
}

func (d *groupConnectors) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_group_connectors"
}

func (d *groupConnectors) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.GroupConnectorsDatasource()
}

func (d *groupConnectors) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.GroupConnectors

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	svc := d.GetClient().NewGroupListConnectors().GroupID(data.Id.ValueString())

	if !data.Schema.IsNull() {
		svc.Schema(data.Schema.ValueString())		
	}

	groupConnectorsResponse, err := svc.Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, groupConnectorsResponse.Code, groupConnectorsResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, groupConnectorsResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
