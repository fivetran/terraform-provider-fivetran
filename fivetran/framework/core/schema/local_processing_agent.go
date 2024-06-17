package schema

import (
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func localDataProcessingAgentAttribute() core.Schema {
    result := core.Schema{
        Fields: map[string]core.SchemaField{
            "id": {
                IsId:        true,
                ValueType:   core.String,
                Description: "The unique identifier for the local processing agent within your account.",
            },
            "group_id": {
                Required:    true,
                ForceNew:    true,
                ValueType:   core.String,
                Description: "The unique identifier for the Group within the Fivetran system.",
            },
            "display_name": {
                Required:    true,
                ForceNew:    true,
                ValueType:   core.String,
                Description: "The unique name for the local processing agent.",
            },
            "registered_at": {
                Readonly:    true,
                ValueType:   core.String,
                Description: "The timestamp of the time the local processing agent was created in your account.",
            },
            "authentication_counter": {
                ResourceOnly: true,
                ValueType:    core.Integer,
                Description:  "Determines whether re-authentication needs to be performed.",
            },
            "config_json": {
                ResourceOnly: true,
                Readonly:     true,
                ValueType:    core.String,
                Description:  "Base64-encoded content of the config.json file.",
            },
            "auth_json": {
                ResourceOnly: true,
                Readonly:     true,
                ValueType:    core.String,
                Description:  "Base64-encoded content of the auth.json file.",
            },
            "docker_compose_yaml": {
                ResourceOnly: true,
                Readonly:     true,
                ValueType:    core.String,
                Description:  "Base64-encoded content of the compose file for the chosen containerization type.",
            },
        },
    }
    return result
}

func localDataProcessingAgentUsageAttribute() core.Schema {
    return core.Schema{
        Fields: map[string]core.SchemaField{
            "connection_id": {
                Readonly:    true,
                ValueType:   core.String,
                Description: "The unique identifier of the connection associated with the agent.",
            },
            "schema": {
                Required:    true,
                ValueType:   core.String,
                Description: "The connection schema name.",
            },
            "service": {
                Required:    true,
                ValueType:   core.String,
                Description: "The connection type.",
            },
        },
    }
}

func localDataProcessingAgentDatasourceSchema() map[string]datasourceSchema.Attribute {
    schema := localDataProcessingAgentAttribute().GetDatasourceSchema()

    schema["usage"] = datasourceSchema.SetNestedAttribute{
                        Computed: true,
                        NestedObject: datasourceSchema.NestedAttributeObject{
                            Attributes: localDataProcessingAgentUsageAttribute().GetDatasourceSchema(),
                        },
                    }
    return schema
}

func localDataProcessingAgentResourceSchema() map[string]resourceSchema.Attribute {
    schema := localDataProcessingAgentAttribute().GetResourceSchema()

    schema["usage"] = resourceSchema.SetNestedAttribute{
                        Computed: true,
                        NestedObject: resourceSchema.NestedAttributeObject{
                            Attributes: localDataProcessingAgentUsageAttribute().GetResourceSchema(),
                        },
                    }
    return schema
}

func LocalProcessingAgentResource() resourceSchema.Schema {
    return resourceSchema.Schema{Attributes: localDataProcessingAgentResourceSchema(),}
}

func LocalProcessingAgentDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema{Attributes: localDataProcessingAgentDatasourceSchema(),}
}

func LocalProcessingAgentsDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema{
        Attributes: map[string]datasourceSchema.Attribute{
            "items": datasourceSchema.SetNestedAttribute{
                Computed: true,
                NestedObject: datasourceSchema.NestedAttributeObject{
                    Attributes: localDataProcessingAgentDatasourceSchema(),
                },
            },
        },
    }
}