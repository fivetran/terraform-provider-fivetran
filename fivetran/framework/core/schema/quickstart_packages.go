package schema

import (
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func QuickstartPackagesDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Blocks: map[string]datasourceSchema.Block{
            "packages": datasourceSchema.ListNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "id": datasourceSchema.StringAttribute{
                            Computed:      true,
                            Description: "The unique identifier for the Quickstart transformation package definition within the Fivetran system",
                        },
                        "name": datasourceSchema.StringAttribute{
                            Computed:      true,
                            Description: "The Quickstart transformation package name",
                        },
                        "version": datasourceSchema.StringAttribute{
                            Computed:      true,
                            Description: "The Quickstart package definition version",
                        },
                        "connector_types": datasourceSchema.SetAttribute{
                            Computed:      true,
                            Description: "The set of connector types",
                            ElementType: basetypes.StringType{},
                        },
                        "output_model_names": datasourceSchema.SetAttribute{
                            Computed:      true,
                            Description: "The list of transformation output models",
                            ElementType: basetypes.StringType{},
                        },
                    },
                },
            },
        },
    }
}

func QuickstartPackageDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "id": datasourceSchema.StringAttribute{
                Required:     true,
                Description: "The unique identifier for the Quickstart transformation package definition within the Fivetran system",
            },
            "name": datasourceSchema.StringAttribute{
                Computed:      true,
                Description: "The Quickstart transformation package name",
            },
            "version": datasourceSchema.StringAttribute{
                Computed:      true,
                Description: "The Quickstart package definition version",
            },
            "connector_types": datasourceSchema.SetAttribute{
                Computed:      true,
                Description: "The set of connector types",
                ElementType: basetypes.StringType{},
            },
            "output_model_names": datasourceSchema.SetAttribute{
                Computed:      true,
                Description: "The list of transformation output models",
                ElementType: basetypes.StringType{},
            },
        },
    }
}