package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

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
			"models": DbtModelsNestedDatasourceBlock(),
		},
	}
}

func DbtModelSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The unique identifier for the dbt Model within the Fivetran system.",
			},
			"model_name": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The dbt Model name.",
			},
			"scheduled": {
				ValueType:   core.Boolean,
				Readonly:    true,
				Description: "Boolean specifying whether the model is selected for execution in the dashboard.",
			},
		},
	}
}

func DbtModelsNestedDatasourceBlock() datasourceSchema.SetNestedBlock {
	return datasourceSchema.SetNestedBlock{
		NestedObject: datasourceSchema.NestedBlockObject{
			Attributes: DbtModelSchema().GetDatasourceSchema(),
		},
	}
}
