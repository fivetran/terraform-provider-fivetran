package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ConnectorDestinationSchemaAttribute() resourceSchema.Attribute {
	return resourceSchema.StringAttribute{
		Required: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Description: "The destination schema identifier. " +
			"For connectors with a single schema use just the schema name (e.g. `my_schema`). " +
			"For connectors that require a table, separate schema and table with a dot (e.g. `my_schema.my_table`). " +
			"The provider automatically determines the correct API fields (`schema`/`schema_prefix`, `table`/`table_group_name`) by trying the API.",
	}
}

func ConnectorV2ResourceBlocks(ctx context.Context) map[string]resourceSchema.Block {
	return map[string]resourceSchema.Block{
		"auth": resourceSchema.SingleNestedBlock{
			Attributes: GetResourceAuthSchemaAttributes(),
			Blocks:     GetResourceAuthSchemaBlocks(),
		},
		"timeouts": timeouts.Block(ctx, timeouts.Opts{
			Create: true,
			Update: true,
		}),
	}
}
