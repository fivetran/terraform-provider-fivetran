package schema

import datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

func DbtModelsDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The ID of this datasource (equals to `project_id`).",
			},
			"project_id": datasourceSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the dbt Project within the Fivetran system.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"models": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the dbt Model within the Fivetran system.",
						},
						"model_name": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The dbt Model name.",
						},
						"scheduled": datasourceSchema.BoolAttribute{
							Computed:    true,
							Description: "Boolean specifying whether the model is selected for execution in the dashboard.",
						},
					},
				},
			},
		},
	}
}
