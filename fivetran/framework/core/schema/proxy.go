package schema

import (
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func ProxyResource() resourceSchema.Schema {
    return resourceSchema.Schema {
        Attributes: map[string]resourceSchema.Attribute{
            "id": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The unique identifier for the proxy within your account.",
            },
            "group_region": resourceSchema.StringAttribute{
                Required:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
                Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
            },
            "display_name": resourceSchema.StringAttribute{
                Required:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
                Description: "Proxy agent name.",
            },
            "proxy_server_uri": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The proxy server URI.",
            },      
            "account_id": resourceSchema.StringAttribute{
                Computed:    true,
                Optional:    true,
                Description: "The unique identifier for the account.",
            },
            "registred_at": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The timestamp of the time the proxy agent was created in your account.",
            },
            "token": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The auth token.",
            },
            "salt": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The salt.",
            },
            "created_by": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The actor who created the proxy agent.",
            },
        },
    }
}

func ProxyDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "id": datasourceSchema.StringAttribute{
                Required:    true,
                Description: "The unique identifier for the proxy within your account.",
            },
            "group_region": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
            },
            "display_name": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "Proxy agent name.",
            },
            "proxy_server_uri": resourceSchema.StringAttribute{
                Optional:    true,
                Computed:    true,
                Description: "The proxy server URI.",
            },  
            "account_id": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The unique identifier for the account.",
            },
            "registred_at": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The timestamp of the time the proxy agent was created in your account.",
            },
            "token": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The auth token.",
            },
            "salt": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The salt.",
            },
            "created_by": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The actor who created the proxy agent.",
            },
        },
    }
}


func ProxiesDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Blocks: map[string]datasourceSchema.Block{
            "items": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier for the proxy within your account.",
                        },
                        "account_id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier for the account.",
                        },
                        "registred_at": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The timestamp of the time the proxy agent was created in your account.",
                        },
                        "region": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
                        },
                        "token": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The auth token.",
                        },
                        "salt": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The salt.",
                        },
                        "created_by": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The actor who created the proxy agent.",
                        },
                        "display_name": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "Proxy agent name.",
                        },
                    },
                },
            },
        },
    }
}

func ProxyConnectionsDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "proxy_id": datasourceSchema.StringAttribute{
                Required:      true,
                Description:   "The ID of this resource.",
            },
        },
        Blocks: map[string]datasourceSchema.Block{
            "connections": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "connection_id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier for the connection within your account.",
                        },
                    },
                },
            },
        },
    }
}
