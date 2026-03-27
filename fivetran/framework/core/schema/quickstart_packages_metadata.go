package schema

import (
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func configurableVarsSchema() datasourceSchema.MapNestedAttribute {
	return datasourceSchema.MapNestedAttribute{
		Computed:    true,
		Description: "Map of configurable variable definitions for the package, keyed by variable name.",
		NestedObject: datasourceSchema.NestedAttributeObject{
			Attributes: map[string]datasourceSchema.Attribute{
				"type": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "The variable type (e.g. STRING, INTEGER, BOOLEAN, DATE).",
				},
				"description": datasourceSchema.StringAttribute{
					Computed:    true,
					Description: "Human-readable description of the variable.",
				},
				"allowed_values": datasourceSchema.ListAttribute{
					Computed:    true,
					Description: "List of allowed values for the variable, if restricted.",
					ElementType: basetypes.StringType{},
				},
			},
		},
	}
}

func QuickstartPackagesDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Blocks: map[string]datasourceSchema.Block{
			"packages": datasourceSchema.ListNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the Quickstart transformation package definition within the Fivetran system",
						},
						"name": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The Quickstart transformation package name",
						},
						"version": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The Quickstart package definition version",
						},
						"connector_types": datasourceSchema.SetAttribute{
							Computed:    true,
							Description: "The set of connector types",
							ElementType: basetypes.StringType{},
						},
						"output_model_names": datasourceSchema.SetAttribute{
							Computed:    true,
							Description: "The list of transformation output models",
							ElementType: basetypes.StringType{},
						},
						"configurable_variables": configurableVarsSchema(),
					},
				},
			},
		},
	}
}

func QuickstartPackageDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the Quickstart transformation package definition within the Fivetran system",
			},
			"name": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The Quickstart transformation package name",
			},
			"version": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The Quickstart package definition version",
			},
			"connector_types": datasourceSchema.SetAttribute{
				Computed:    true,
				Description: "The set of connector types",
				ElementType: basetypes.StringType{},
			},
			"output_model_names": datasourceSchema.SetAttribute{
				Computed:    true,
				Description: "The list of transformation output models",
				ElementType: basetypes.StringType{},
			},
			"configurable_variables": configurableVarsSchema(),
		},
	}
}
