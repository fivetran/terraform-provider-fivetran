package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	sdk "github.com/fivetran/go-fivetran/groups"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func GroupConnections() datasource.DataSource {
	return &groupConnections{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &groupConnections{}

type groupConnections struct {
	core.ProviderDatasource
}

func (d *groupConnections) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_group_connections"
}

func (d *groupConnections) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.GroupConnectionsDatasource()
}

func (d *groupConnections) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.GroupConnections
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var listResponse sdk.GroupListConnectionsResponse
	limit := 1000

	for {
		var err error
		var tmpResp sdk.GroupListConnectionsResponse
		svc := d.GetClient().NewGroupListConnections()
		
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
			listResponse = sdk.GroupListConnectionsResponse{}
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
