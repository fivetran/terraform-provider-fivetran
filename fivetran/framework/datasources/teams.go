package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
    sdk "github.com/fivetran/go-fivetran/teams"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func Teams() datasource.DataSource {
	return &teams{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &teams{}

type teams struct {
	core.ProviderDatasource
}

func (d *teams) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_teams"
}

func (d *teams) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.TeamsDatasource()
}

func (d *teams) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Teams
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var listResponse sdk.TeamsListResponse
	limit := 1000

	for {
		var err error
		var tmpResp sdk.TeamsListResponse
		svc := d.GetClient().NewTeamsList()
		
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
			listResponse = sdk.TeamsListResponse{}
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
