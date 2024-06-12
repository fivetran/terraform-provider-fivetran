package datasources

import (
    "context"
    "fmt"

    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
    "github.com/hashicorp/terraform-plugin-framework/datasource"

    fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func ProxyAgent() datasource.DataSource {
    return &proxyAgent{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &proxyAgent{}

type proxyAgent struct {
    core.ProviderDatasource
}

func (d *proxyAgent) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = "fivetran_proxy_agent"
}

func (d *proxyAgent) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = fivetranSchema.ProxyAgentDatasource()
}

func (d *proxyAgent) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    if d.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.ProxyAgentDatasourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    detailsResponse, err := d.GetClient().NewProxyDetails().ProxyId(data.Id.ValueString()).Do(ctx)

    if err != nil {
        resp.Diagnostics.AddError(
            "Read error.",
            fmt.Sprintf("%v; code: %v", err, detailsResponse.Code),
        )
        return
    }

    data.ReadFromResponse(detailsResponse)

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}