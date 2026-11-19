package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TransformationProject() datasource.DataSource {
	return &transformationProject{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &transformationProject{}

type transformationProject struct {
	core.ProviderDatasource
}

func (d *transformationProject) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_transformation_project"
}

func (d *transformationProject) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.TransformationProjectDatasource()
}

func (d *transformationProject) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TransformationDatasourceProject

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	projectResponse, err := d.GetClient().NewTransformationProjectDetails().ProjectId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"TransformationProject Read Error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, projectResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}