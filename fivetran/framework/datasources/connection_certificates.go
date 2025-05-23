package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func ConnectionCertificates() datasource.DataSource {
	return &connectionCertificates{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &connectionCertificates{}

type connectionCertificates struct {
	core.ProviderDatasource
}

func (d *connectionCertificates) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_connection_certificates"
}

func (d *connectionCertificates) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.CertificateConnectionDatasource()
}

func (d *connectionCertificates) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.CertificatesConnection
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	listResponse, err := core.ReadCertificatesFromUpstream(ctx, d.GetClient(), data.Id.ValueString(), "connection")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Connection Certificates DataSource.",
			fmt.Sprintf("%v; code: %v", err, listResponse.Code),
		)

		return
	}

	data.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
