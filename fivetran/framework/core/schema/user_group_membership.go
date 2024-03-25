package schema

import (
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func UserGroupMembershipResource() resourceSchema.Schema {
    return resourceSchema.Schema {
        Attributes: map[string]resourceSchema.Attribute{
            "user_id": resourceSchema.StringAttribute{
                Required:    true,
                Description: "The unique identifier for the user within your account.",
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
                            Description: "The user's role that links the user and the group",
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

func UserGroupMembershipDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "user_id": datasourceSchema.StringAttribute{
                Required:    true,
                Description: "The unique identifier for the user within your account.",
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
                            Description: "The user's role that links the user and the group",
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