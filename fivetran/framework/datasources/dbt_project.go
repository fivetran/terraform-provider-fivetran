package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/dbt"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func DbtProject() datasource.DataSource {
	return &dbtProject{}
}

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSourceWithConfigure = &dbtProject{}

type dbtProject struct {
	core.ProviderDatasource
}

func (d *dbtProject) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "fivetran_dbt_project"
}

func (d *dbtProject) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fivetranSchema.DbtProjectDatasource()
}

func (d *dbtProject) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtProject

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	projectResponse, err := d.GetClient().NewDbtProjectDetails().DbtProjectID(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"DbtProject Read Error.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, projectResponse, nil)

	if strings.ToLower(projectResponse.Data.Status) == "ready" {
		modelsResp, err := GetAllDbtModelsForProject(d.GetClient(), ctx, projectResponse.Data.ID, 1000)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"DbtProject Models Read Error.",
				fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
			)
		} else {
			data.ReadFromResponse(ctx, projectResponse, &modelsResp)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func GetAllDbtModelsForProject(client *fivetran.Client, ctx context.Context, projectId string, limit int) (dbt.DbtModelsListResponse, error) {
	var resp dbt.DbtModelsListResponse
	var respNextCursor string

	for {
		var err error
		var respInner dbt.DbtModelsListResponse
		svc := client.NewDbtModelsList().ProjectId(projectId)
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return dbt.DbtModelsListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
