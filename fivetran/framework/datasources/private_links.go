package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
    sdk "github.com/fivetran/go-fivetran/private_link"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func PrivateLinks() datasource.DataSource {
	return &privateLinks{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &privateLinks{}

type privateLinks struct {
	core.ProviderDatasource
}

func (d *privateLinks) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_private_links"
}

func (d *privateLinks) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.PrivateLinksDatasource()
}

func (d *privateLinks) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.PrivateLinks
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var listResponse sdk.PrivateLinkListResponse
	limit := 1000

	for {
		var err error
		var tmpResp sdk.PrivateLinkListResponse
		svc := d.GetClient().NewPrivateLinkList()
		
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
			listResponse = sdk.PrivateLinkListResponse{}
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