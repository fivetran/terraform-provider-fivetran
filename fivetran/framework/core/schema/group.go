package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func GroupSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the group within the Fivetran system.",
			},
			"name": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The name of the group within your account.",
			},
			"created_at": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The timestamp of when the group was created in your account.",
			},
			"last_updated": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The timestamp of when the resource/datasource was updated last time.",
			},
		},
	}
}

func GroupResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: GroupSchema().GetResourceSchema(),
	}
}

func GroupDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: GroupSchema().GetDatasourceSchema(),
	}
}

func GroupsDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"groups": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: GroupSchema().GetDatasourceSchema(),
				},
			},
		},
	}
}
