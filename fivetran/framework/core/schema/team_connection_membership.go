package schema

import (
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TeamConnectionMembershipDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "id": datasourceSchema.StringAttribute{
                Computed:      true,
                Description:   "The unique identifier for the team within your account.",
            },
        },
        Blocks: map[string]datasourceSchema.Block{
            "connection": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "connection_id": datasourceSchema.StringAttribute{
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