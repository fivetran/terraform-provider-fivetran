package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func ExternalLogging() datasource.DataSource {
	return &externalLogging{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &externalLogging{}

type externalLogging struct {
	core.ProviderDatasource
}

func (d *externalLogging) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_external_logging"
}

func (d *externalLogging) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.ExternalLoggingDatasource()
}

func (d *externalLogging) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.ExternalLogging

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	listResponse, err := d.GetClient().NewExternalLoggingDetails().ExternalLoggingId(data.Id.ValueString()).DoCustom(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v", err, listResponse.Code),
		)
		return
	}

	data.ReadFromCustomResponse(ctx, listResponse, nil)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
