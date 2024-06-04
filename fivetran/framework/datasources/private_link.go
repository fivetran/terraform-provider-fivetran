package datasources

import (
    "context"
    "fmt"

    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
    "github.com/hashicorp/terraform-plugin-framework/datasource"

    fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func PrivateLink() datasource.DataSource {
    return &privateLink{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &privateLink{}

type privateLink struct {
    core.ProviderDatasource
}

func (d *privateLink) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = "fivetran_private_link"
}

func (d *privateLink) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = fivetranSchema.PrivateLinkDatasource()
}

func (d *privateLink) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    if d.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.PrivateLink

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    detailsResponse, err := d.GetClient().NewPrivateLinkDetails().PrivateLinkId(data.Id.ValueString()).Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Read error.",
            fmt.Sprintf("%v; code: %v", err, detailsResponse.Code),
        )
        return
    }

    data.ReadFromResponse(ctx, detailsResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}