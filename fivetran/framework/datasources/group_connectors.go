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

	var respNextCursor string
	var listResponse model.GroupConnectors
	limit := 1000

	for {
		var err error
		var tmpResp model.GroupConnectors
		svc := d.GetClient().NewGroupListConnectors()
		
		if respNextCursor == "" {
			tmpResp, err = svc.Limit(limit).GroupID(data.Id.ValueString()).Do(ctx)
		}

		if respNextCursor != "" {
			tmpResp, err = svc.Limit(limit).GroupID(data.Id.ValueString()).Cursor(respNextCursor).Do(ctx)
		}
		
		if err != nil {
			resp.Diagnostics.AddError(
				"Read error.",
				fmt.Sprintf("%v; code: %v; message: %v", err, tmpResp.Code, tmpResp.Message),
			)
			listResponse = sdk.GroupConnectors{}
		}

		listResponse.Data.Items = append(listResponse.Data.Items, tmpResp.Data.Items...)

		if tmpResp.Data.NextCursor == "" {
			break
		}

		respNextCursor = tmpResp.Data.NextCursor
	}

	data.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
