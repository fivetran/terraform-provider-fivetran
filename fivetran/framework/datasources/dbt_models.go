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

func DbtModels() datasource.DataSource {
	return &dbtModels{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &dbtModels{}

type dbtModels struct {
	core.ProviderDatasource
}

func (d *dbtModels) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_dbt_models"
}

func (d *dbtModels) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.DbtModelsDatasource()
}

func (d *dbtModels) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtModels
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var dbtModelsResponse dbt.DbtModelsListResponse
	limit := 1000

	for {
		var err error
		var tmpResp dbt.DbtModelsListResponse
		svc := d.GetClient().NewDbtModelsList()
		svc.ProjectId(data.ProjectId.ValueString())

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
			dbtModelsResponse = dbt.DbtModelsListResponse{}
		}

		dbtModelsResponse.Data.Items = append(dbtModelsResponse.Data.Items, tmpResp.Data.Items...)

		if tmpResp.Data.NextCursor == "" {
			break
		}

		respNextCursor = tmpResp.Data.NextCursor
	}

	data.ReadFromResponse(ctx, dbtModelsResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
