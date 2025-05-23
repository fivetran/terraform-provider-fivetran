package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
    webhooksSdk "github.com/fivetran/go-fivetran/webhooks"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func Webhooks() datasource.DataSource {
	return &webhooks{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &webhooks{}

type webhooks struct {
	core.ProviderDatasource
}

func (d *webhooks) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_webhooks"
}

func (d *webhooks) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.WebhooksDatasource()
}

func (d *webhooks) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.Webhooks
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var webhookResponse webhooksSdk.WebhookListResponse
	limit := 1000

	for {
		var err error
		var tmpResp webhooksSdk.WebhookListResponse
		svc := d.GetClient().NewWebhookList()
		
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
			webhookResponse = webhooksSdk.WebhookListResponse{}
		}

		webhookResponse.Data.Items = append(webhookResponse.Data.Items, tmpResp.Data.Items...)

		if tmpResp.Data.NextCursor == "" {
			break
		}

		respNextCursor = tmpResp.Data.NextCursor
	}

	data.ReadFromResponse(ctx, webhookResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
