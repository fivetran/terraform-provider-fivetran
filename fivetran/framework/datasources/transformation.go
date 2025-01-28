package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func Transformation() datasource.DataSource {
	return &transformation{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &transformation{}

type transformation struct {
	core.ProviderDatasource
}

func (d *transformation) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_transformation"
}

func (d *transformation) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.TransformationDatasource()
}

func (d *transformation) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Transformation

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	readResponse, err := d.GetClient().NewTransformationDetails().TransformationId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Transformation Read Error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, readResponse.Code, readResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, readResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}