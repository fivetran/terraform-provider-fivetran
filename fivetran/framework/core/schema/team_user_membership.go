package schema

import (
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TeamUserMembershipResource() resourceSchema.Schema {
    return resourceSchema.Schema {
        Attributes: map[string]resourceSchema.Attribute{
            /*"id": resourceSchema.StringAttribute{
                Computed:      true,
                Description:   "The webhook ID",
            },*/
            "team_id": resourceSchema.StringAttribute{
                Required:    true,
                Description: "The unique identifier for the team within your account.",
            },
        },
        Blocks: map[string]resourceSchema.Block{
            "user": resourceSchema.SetNestedBlock{
                NestedObject: resourceSchema.NestedBlockObject{
                    Attributes: map[string]resourceSchema.Attribute{
                        "user_id": resourceSchema.StringAttribute{
                            Required:    true,
                            Description: "The user unique identifier",
                        },
                        "role": resourceSchema.StringAttribute{
                            Required:    true,
                            Description: "The team's role that links the team and the user",
                        },
                    },
                },
            },
        },
    }
}

func TeamUserMembershipDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "team_id": datasourceSchema.StringAttribute{
                Required:    true,
                Description: "The unique identifier for the team within your account.",
            },
        },
        Blocks: map[string]datasourceSchema.Block{
            "user": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "user_id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The user unique identifier",
                        },
                        "role": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The team's role that links the team and the user",
                        },
                    },
                },
            },
        },
    }
}