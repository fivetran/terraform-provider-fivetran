package datasources

import (
	"context"
	"fmt"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
    sdk "github.com/fivetran/go-fivetran/hybrid_deployment_agent"

	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
)

func HybridDeploymentAgents() datasource.DataSource {
	return &hybridDeploymentAgents{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &hybridDeploymentAgents{}

type hybridDeploymentAgents struct {
	core.ProviderDatasource
}

func (d *hybridDeploymentAgents) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_hybrid_deployment_agents"
}

func (d *hybridDeploymentAgents) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.HybridDeploymentAgentsDatasource()
}

func (d *hybridDeploymentAgents) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.HybridDeploymentAgents
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	var respNextCursor string
	var listResponse sdk.HybridDeploymentAgentListResponse
	limit := 1000

	for {
		var err error
		var tmpResp sdk.HybridDeploymentAgentListResponse
		svc := d.GetClient().NewHybridDeploymentAgentList()
		
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
			listResponse = sdk.HybridDeploymentAgentListResponse{}
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
