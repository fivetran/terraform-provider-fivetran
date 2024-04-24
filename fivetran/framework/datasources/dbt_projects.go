package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran/dbt"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func DbtProjects() datasource.DataSource {
	return &dbtProjects{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &dbtProjects{}

type dbtProjects struct {
	core.ProviderDatasource
}

func (d *dbtProjects) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_dbt_projects"
}

func (d *dbtProjects) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.DbtProjectsSchema()
}

func (d *dbtProjects) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtProjects
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var dbtProjectsResponse dbt.DbtProjectsListResponse
	limit := 1000

	for {
		var err error
		var tmpResp dbt.DbtProjectsListResponse
		svc := d.GetClient().NewDbtProjectsList()

		if respNextCursor == "" {
			tmpResp, err = svc.Limit(limit).Do(ctx)
		}

		if respNextCursor != "" {
			tmpResp, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}

		if err != nil {
			resp.Diagnostics.AddError(
				"Read error.",
				fmt.Sprintf("%v; code: %v", err, tmpResp.Code),
			)
			dbtProjectsResponse = dbt.DbtProjectsListResponse{}
		}

		dbtProjectsResponse.Data.Items = append(dbtProjectsResponse.Data.Items, tmpResp.Data.Items...)

		if tmpResp.Data.NextCursor == "" {
			break
		}

		respNextCursor = tmpResp.Data.NextCursor
	}

	data.ReadFromResponse(ctx, dbtProjectsResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
