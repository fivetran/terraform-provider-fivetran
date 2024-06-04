package datasources

import (
    "context"
    "fmt"

    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
    "github.com/hashicorp/terraform-plugin-framework/datasource"

    fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func Proxy() datasource.DataSource {
    return &proxy{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &proxy{}

type proxy struct {
    core.ProviderDatasource
}

func (d *proxy) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = "fivetran_proxy"
}

func (d *proxy) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = fivetranSchema.ProxyDatasource()
}

func (d *proxy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    if d.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.Proxy

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    detailsResponse, err := d.GetClient().NewProxyDetails().ProxyId(data.Id.ValueString()).Do(ctx)

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