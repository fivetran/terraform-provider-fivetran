package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
    sdk "github.com/fivetran/go-fivetran/fingerprints"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func ConnectorFingerprints() datasource.DataSource {
	return &connectorFingerprints{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &connectorFingerprints{}

type connectorFingerprints struct {
	core.ProviderDatasource
}

func (d *connectorFingerprints) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_connector_fingerprints"
}

func (d *connectorFingerprints) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.FingerprintConnectorDatasource()
}

func (d *connectorFingerprints) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.FingerprintsConnector
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var listResponse sdk.FingerprintsListResponse
	limit := 1000

	for {
		var err error
		var tmpResp sdk.FingerprintsListResponse
		svc := d.GetClient().NewConnectorFingerprintsList().ConnectorID(data.Id.ValueString())
		
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
			listResponse = sdk.FingerprintsListResponse{}
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
