package datasources

import (
	"context"
	"fmt"

	sdk "github.com/fivetran/go-fivetran/transformations"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TransformationProjects() datasource.DataSource {
	return &transformationProjects{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &transformationProjects{}

type transformationProjects struct {
	core.ProviderDatasource
}

func (d *transformationProjects) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_transformation_projects"
}

func (d *transformationProjects) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.TransformationProjectListDatasource()
}

func (d *transformationProjects) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TransformationProjects
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var listResponse sdk.TransformationProjectsListResponse
	limit := 1000

	for {
		var err error
		var tmpResp sdk.TransformationProjectsListResponse
		svc := d.GetClient().NewTransformationProjectsList()

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
			listResponse = sdk.TransformationProjectsListResponse{}
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
