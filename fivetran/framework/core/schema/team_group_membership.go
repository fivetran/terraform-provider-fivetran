package schema

import (
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TeamGroupMembershipResource() resourceSchema.Schema {
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
            "group": resourceSchema.SetNestedBlock{
                NestedObject: resourceSchema.NestedBlockObject{
                    Attributes: map[string]resourceSchema.Attribute{
                        "group_id": resourceSchema.StringAttribute{
                            Required:    true,
                            Description: "The group unique identifier",
                        },
                        "role": resourceSchema.StringAttribute{
                            Required:    true,
                            Description: "The team's role that links the team and the group",
                        },
                        "created_at": resourceSchema.StringAttribute{
                            Computed:      true,
                            Description:   "The date and time the membership was created",
                        },
                    },
                },
            },
        },
    }
}

func TeamGroupMembershipDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "team_id": datasourceSchema.StringAttribute{
                Required:    true,
                Description: "The unique identifier for the team within your account.",
            },
        },
        Blocks: map[string]datasourceSchema.Block{
            "group": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "group_id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The group unique identifier",
                        },
                        "role": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The team's role that links the team and the group",
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