package schema

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/fivetrantypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
				Optional:    true,
				Computed:    true,
				Description:   "The unique identifier for the connector within the Fivetran system.",
			},
			"group_id": schema.StringAttribute{
				Optional: true,
				Description:   "The unique identifier for the Group (Destination) within the Fivetran system.",
			},
			"connector_name": schema.StringAttribute{
				Optional: true,
				Description:   "The name used both as the connection's name within the Fivetran system and as the source schema's name within your destination.",
			},
			"schema_change_handling": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf("ALLOW_ALL", "ALLOW_COLUMNS", "BLOCK_ALL"),
				},
				Description: "The value specifying how new source data is handled.",
			},
			"validation_level": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("TABLES"),
				Validators: []validator.String{
					stringvalidator.OneOf("NONE", "TABLES", "COLUMNS"),
				},
				Description: `
The value defines validation method. 
- NONE: no validation, any configuration accepted. 
- TABLES: validate table names, fail on attempt to configure non-existing schemas/tables.
- COLUMNS: validate the whole schema config including column names. The resource will try to fetch columns for every configured table and verify column names.
`,
			},
			"schemas": schema.MapNestedAttribute{
				Optional:    true,
				Description: "Map of schema configurations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The boolean value specifying whether the sync for the schema into the destination is enabled.",
						},
						"tables": schema.MapNestedAttribute{
							Description: "Map of table configurations.",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Description: "The boolean value specifying whether the sync for the table into the destination is enabled.",
									},
									"sync_mode": schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: "This field appears in the response if the connector supports switching sync modes for tables.",
										Validators: []validator.String{
											stringvalidator.OneOf("HISTORY", "SOFT_DELETE", "LIVE"),
										},
									},
									"columns": schema.MapNestedAttribute{
										Description: "Map of table configurations.",
										Optional:    true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
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
												"is_primary_key": schema.BoolAttribute{
													Optional:    true,
													Computed:    true,
													Description: "",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"schemas_json": schema.StringAttribute{
				Optional:    true,
				CustomType:  fivetrantypes.JsonSchemaType{},
				Description: "Schema settings in Json format, following Fivetran API endpoint contract for `schemas` field (a map of schemas).",
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
		DeprecationMessage: "Configure `schemas` instead. This attribute will be removed in the next major version of the provider.",
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
				"is_primary_key": schema.BoolAttribute{
					Optional:    true,
					Description: "",
				},
			},
		},
	}
}
