package schema

import (
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func LocalProcessingAgentResource() resourceSchema.Schema {
    return resourceSchema.Schema {
        Attributes: map[string]resourceSchema.Attribute{
            "id": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The unique identifier for the local processing agent within your account.",
            },
            "group_id": resourceSchema.StringAttribute{
                Required:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
                Description: "The unique identifier for the group or group within the Fivetran system.",
            },
            "display_name": resourceSchema.StringAttribute{
                Required:    true,
                PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
                Description: "The unique name for the local processing agent.",
            },
            "re_auth": resourceSchema.BoolAttribute{
                Optional:    true,
                Description: "Determines whether re-authentication needs to be performed.",
            },
            "registered_at": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "The timestamp of the time the local processing agent was created in your account.",
            },
            "config_json": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "Base64-encoded content of the config.json file.",
            },
            "auth_json": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "Base64-encoded content of the auth.json file.",
            },
            "docker_compose_yaml": resourceSchema.StringAttribute{
                Computed:    true,
                Description: "Base64-encoded content of the compose file for the chosen containerization type.",
            },
        },
        Blocks: map[string]resourceSchema.Block{
            "usage": resourceSchema.SetNestedBlock{
                NestedObject: resourceSchema.NestedBlockObject{
                    Attributes: map[string]resourceSchema.Attribute{
                        "connection_id": resourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier of the connection associated with the agent.",
                        },
                        "schema": resourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The connection schema name.",
                        },
                        "service": resourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The connection type.",
                        },
                    },
                },
            },
        },
    }
}

func LocalProcessingAgentDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "id": datasourceSchema.StringAttribute{
                Required:    true,
                Description: "The unique identifier for the local processing agent within your account.",
            },
            "group_id": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The unique identifier for the group or group within the Fivetran system.",
            },
            "display_name": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The unique name for the local processing agent.",
            },
            "registered_at": datasourceSchema.StringAttribute{
                Computed:    true,
                Description: "The timestamp of the time the local processing agent was created in your account.",
            },
            "re_auth": datasourceSchema.BoolAttribute{
                Optional:    true,
                Description: "Determines whether re-authentication needs to be performed.",
            },
            "config_json": datasourceSchema.StringAttribute{
                Optional:    true,
                Computed:    true,
                Description: "Base64-encoded content of the config.json file.",
            },
            "auth_json": datasourceSchema.StringAttribute{
                Optional:    true,
                Computed:    true,
                Description: "Base64-encoded content of the auth.json file.",
            },
            "docker_compose_yaml": datasourceSchema.StringAttribute{
                Optional:    true,
                Computed:    true,
                Description: "Base64-encoded content of the compose file for the chosen containerization type.",
            },
        },
        Blocks: map[string]datasourceSchema.Block{
            "usage": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "connection_id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier of the connection associated with the agent.",
                        },
                        "schema": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The connection schema name.",
                        },
                        "service": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The connection type.",
                        },
                    },
                },
            },
        },
    }
}

func LocalProcessingAgentsDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema {
        Blocks: map[string]datasourceSchema.Block{
            "items": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
                        "id": datasourceSchema.StringAttribute{
                            Required:    true,
                            Description: "The unique identifier for the proxy within your account.",
                        },
                        "group_id": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique identifier for the group or group within the Fivetran system.",
                        },
                        "display_name": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The unique name for the local processing agent.",
                        },
                        "registered_at": datasourceSchema.StringAttribute{
                            Computed:    true,
                            Description: "The timestamp of the time the local processing agent was created in your account.",
                        },
                    },
                    Blocks: map[string]datasourceSchema.Block{
                        "usage": datasourceSchema.SetNestedBlock{
                            NestedObject: datasourceSchema.NestedBlockObject{
                                Attributes: map[string]datasourceSchema.Attribute{
                                    "connection_id": datasourceSchema.StringAttribute{
                                        Computed:    true,
                                        Description: "The unique identifier of the connection associated with the agent.",
                                    },
                                    "schema": datasourceSchema.StringAttribute{
                                        Computed:    true,
                                        Description: "The connection schema name.",
                                    },
                                    "service": datasourceSchema.StringAttribute{
                                        Computed:    true,
                                        Description: "The connection type.",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}
