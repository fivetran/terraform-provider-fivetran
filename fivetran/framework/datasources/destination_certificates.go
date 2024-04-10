package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func DestinationCertificates() datasource.DataSource {
	return &destinationCertificates{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &destinationCertificates{}

type destinationCertificates struct {
	core.ProviderDatasource
}

func (d *destinationCertificates) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_destination_certificates"
}

func (d *destinationCertificates) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.CertificateDestinationDatasource()
}

func (d *destinationCertificates) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.CertificatesDestination
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	listResponse, err := core.ReadCertificatesFromUpstream(ctx, d.GetClient(), data.Id.ValueString(), "destination")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Destination Certificates DataSource.",
			fmt.Sprintf("%v; code: %v", err, listResponse.Code),
		)

		return
	}

	data.ReadFromResponse(ctx, listResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
