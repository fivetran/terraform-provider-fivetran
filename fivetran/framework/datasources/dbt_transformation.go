package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func DbtTransformation() datasource.DataSource {
	return &dbtTransformation{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &dbtTransformation{}

type dbtTransformation struct {
	core.ProviderDatasource
}

func (d *dbtTransformation) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_dbt_transformation"
}

func (d *dbtTransformation) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.DbtTransformationDatasourceSchema()
}

func (d *dbtTransformation) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtTransformation

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	transformationResponse, err := d.GetClient().NewDbtTransformationDetailsService().TransformationId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"DbtTransformation Read Error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, transformationResponse.Code, transformationResponse.Message),
		)
		return
	}

	modelResponse, err := d.GetClient().NewDbtModelDetails().ModelId(transformationResponse.Data.DbtModelId).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"DbtTransformation model Read Error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, transformationResponse.Code, transformationResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, transformationResponse, &modelResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
