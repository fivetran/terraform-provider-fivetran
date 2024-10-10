package schema

import (
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func hybridDeploymentAgentAttribute() core.Schema {
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

func hybridDeploymentAgentUsageAttribute() core.Schema {
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

func hybridDeploymentAgentDatasourceSchema() map[string]datasourceSchema.Attribute {
    schema := hybridDeploymentAgentAttribute().GetDatasourceSchema()

    schema["usage"] = datasourceSchema.SetNestedAttribute{
                        Computed: true,
                        NestedObject: datasourceSchema.NestedAttributeObject{
                            Attributes: hybridDeploymentAgentUsageAttribute().GetDatasourceSchema(),
                        },
                    }
    return schema
}

func hybridDeploymentAgentResourceSchema() map[string]resourceSchema.Attribute {
    schema := hybridDeploymentAgentAttribute().GetResourceSchema()

    schema["usage"] = resourceSchema.SetNestedAttribute{
                        Computed: true,
                        NestedObject: resourceSchema.NestedAttributeObject{
                            Attributes: hybridDeploymentAgentUsageAttribute().GetResourceSchema(),
                        },
                    }
    return schema
}

func HybridDeploymentAgentResource() resourceSchema.Schema {
    return resourceSchema.Schema{Attributes: hybridDeploymentAgentResourceSchema(),}
}

func HybridDeploymentAgentDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema{Attributes: hybridDeploymentAgentDatasourceSchema(),}
}

func HybridDeploymentAgentsDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema{
        Attributes: map[string]datasourceSchema.Attribute{
            "items": datasourceSchema.SetNestedAttribute{
                Computed: true,
                NestedObject: datasourceSchema.NestedAttributeObject{
                    Attributes: hybridDeploymentAgentDatasourceSchema(),
                },
            },
        },
    }
}
