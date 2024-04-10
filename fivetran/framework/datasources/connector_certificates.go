package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func ConnectorCertificates() datasource.DataSource {
	return &connectorCertificates{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &connectorCertificates{}

type connectorCertificates struct {
	core.ProviderDatasource
}

func (d *connectorCertificates) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_connector_certificates"
}

func (d *connectorCertificates) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.CertificateConnectorDatasource()
}

func (d *connectorCertificates) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.CertificatesConnector
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	listResponse, err := core.ReadCertificatesFromUpstream(ctx, d.GetClient(), data.Id.ValueString(), "connector")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Connector Certificates DataSource.",
			fmt.Sprintf("%v; code: %v", err, listResponse.Code),
		)

		return
	}

	data.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
