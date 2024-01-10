package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func GetConnectorSchemaResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique resource identifier (equals to `connector_id`).",
			},
			"connector_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The unique identifier for the connector within the Fivetran system.",
			},
			"schema_change_handling": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("ALLOW_ALL", "ALLOW_COLUMNS", "BLOCK_ALL"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"schema": getSchemaBlock(),
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Read:   true,
				Create: true,
				Update: true,
			}),
		},
	}
}

func getSchemaBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "The schema name within your destination in accordance with Fivetran conventional rules.",
				},
				"enabled": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The boolean value specifying whether the sync for the schema into the destination is enabled.",
				},
			},
			Blocks: map[string]schema.Block{
				"table": getTableBlock(),
			},
		},
	}
}

func getTableBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "The table name within your destination in accordance with Fivetran conventional rules.",
				},
				"enabled": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The boolean value specifying whether the sync of table into the destination is enabled.",
				},
				"sync_mode": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "This field appears in the response if the connector supports switching sync modes for tables.",
					Validators: []validator.String{
						stringvalidator.OneOf("HISTORY", "SOFT_DELETE", "LIVE"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"column": getColumnBlock(),
			},
		},
	}
}

func getColumnBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required:    true,
					Description: "The column name within your destination in accordance with Fivetran conventional rules.",
				},
				"enabled": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The boolean value specifying whether the sync of the column into the destination is enabled.",
				},
				"hashed": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The boolean value specifying whether a column should be hashed.",
				},
			},
		},
	}
}
