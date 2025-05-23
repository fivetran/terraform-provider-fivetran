package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
    sdk "github.com/fivetran/go-fivetran/destinations"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func Destinations() datasource.DataSource {
	return &destinations{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &destinations{}

type destinations struct {
	core.ProviderDatasource
}

func (d *destinations) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_destinations"
}

func (d *destinations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.DestinationsDatasource()
}

func (d *destinations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Destinations
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var destinationsResponse sdk.DestinationsListResponse
	limit := 1000

	for {
		var err error
		var tmpResp sdk.DestinationsListResponse
		svc := d.GetClient().NewDestinationsList()
		
        svc.Limit(limit)
        if respNextCursor != "" {
            svc.Cursor(respNextCursor)
        }
        tmpResp, err = svc.Do(ctx)
		
		if err != nil {
			resp.Diagnostics.AddError(
				"Read error.",
				fmt.Sprintf("%v; code: %v", err, tmpResp.Code),
			)
			destinationsResponse = sdk.DestinationsListResponse{}
		}

		destinationsResponse.Data.Items = append(destinationsResponse.Data.Items, tmpResp.Data.Items...)

		if tmpResp.Data.NextCursor == "" {
			break
		}

		respNextCursor = tmpResp.Data.NextCursor
	}

	data.ReadFromResponse(ctx, destinationsResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}