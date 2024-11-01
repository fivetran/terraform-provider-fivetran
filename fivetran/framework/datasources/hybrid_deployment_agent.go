package datasources

import (
    "context"
    "fmt"

    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
    "github.com/hashicorp/terraform-plugin-framework/datasource"

    fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func HybridDeploymentAgent() datasource.DataSource {
    return &hybridDeploymentAgent{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &hybridDeploymentAgent{}

type hybridDeploymentAgent struct {
    core.ProviderDatasource
}

func (d *hybridDeploymentAgent) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = "fivetran_hybrid_deployment_agent"
}

func (d *hybridDeploymentAgent) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = fivetranSchema.HybridDeploymentAgentDatasource()
}

func (d *hybridDeploymentAgent) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    if d.GetClient() == nil {
        resp.Diagnostics.AddError(
            "Unconfigured Fivetran Client",
            "Please report this issue to the provider developers.",
        )

        return
    }

    var data model.HybridDeploymentAgentDatasourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    detailsResponse, err := d.GetClient().NewHybridDeploymentAgentDetails().AgentId(data.Id.ValueString()).Do(ctx)

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