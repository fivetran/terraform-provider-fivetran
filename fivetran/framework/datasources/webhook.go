package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func Webhook() datasource.DataSource {
	return &webhook{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &webhook{}

type webhook struct {
	core.ProviderDatasource
}

func (d *webhook) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_webhook"
}

func (d *webhook) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.WebhookDatasource()
}

func (d *webhook) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Webhook

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	webhookResponse, err := d.GetClient().NewWebhookDetails().WebhookId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Read error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, webhookResponse.Code, webhookResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, webhookResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
