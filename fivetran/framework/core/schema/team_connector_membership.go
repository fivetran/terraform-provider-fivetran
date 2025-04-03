package schema

import (
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TeamConnectionMembershipResource() resourceSchema.Schema {
    return resourceSchema.Schema {
        Attributes: map[string]resourceSchema.Attribute{
            "id": resourceSchema.StringAttribute{
                Computed:      true,
                Description:   "The unique identifier for resource.",
            },
            "team_id": resourceSchema.StringAttribute{
                Required:       true,
                Description:    "The unique identifier for the team within your account.",
            },
        },
        Blocks: map[string]resourceSchema.Block{
            "connector": resourceSchema.SetNestedBlock{
                NestedObject: resourceSchema.NestedBlockObject{
                    Attributes: map[string]resourceSchema.Attribute{
                        "connector_id": resourceSchema.StringAttribute{
                            Required:    true,
                            Description: "The connection unique identifier",
                        },
                        "role": resourceSchema.StringAttribute{
                            Required:    true,
                            Description: "The team's role that links the team and the connection",
                        },
                        "created_at": resourceSchema.StringAttribute{
                            Computed:       true,
                            Description:    "The date and time the membership was created",
                        },
                    },
                },
            },
        },
    }
}

func TeamConnectionMembershipDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "id": resourceSchema.StringAttribute{
                Computed:      true,
                Description:   "The unique identifier for resource.",
            },
            "team_id": datasourceSchema.StringAttribute{
                Required:    true,
                Description: "The unique identifier for the team within your account.",
            },
        },
        Blocks: map[string]datasourceSchema.Block{
            "connector": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "connector_id": datasourceSchema.StringAttribute{
                            Required:    true,
                            Description: "The connection unique identifier",
                        },
                        "role": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The team's role that links the team and the connection",
                        },
                        "created_at": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The date and time the membership was created",
                        },
                    },
                },
            },
        },
    }
}