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
                Description: "The unique identifier for the hybrid deployment agent within your account.",
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
                Description: "The unique name for the hybrid deployment agent.",
            },
            "auth_type": {
                Required:     true,
                ResourceOnly: true,
                ValueType:    core.String,
                Description:  "Type of authentification. Possible values `AUTO`,`MANUAL`",
            },
            "env_type": {
                Required:     true,
                ResourceOnly: true,
                ValueType:    core.String,
                Description:  "Environment type. Possible values `DOCKER`,`PODMAN`,`KUBERNETES`,`SNOWPARK`",
            },
            "registered_at": {
                Readonly:    true,
                ValueType:   core.String,
                Description: "The timestamp of the time the hybrid deployment agent was created in your account.",
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
            "token": {
                ResourceOnly: true,
                Readonly:     true,
                ValueType:    core.String,
                Description:  "Base64 encoded content of token.",
            },
        },
    }
    return result
}

func HybridDeploymentAgentResource() resourceSchema.Schema {
    return resourceSchema.Schema{Attributes: hybridDeploymentAgentAttribute().GetResourceSchema(),}
}

func HybridDeploymentAgentDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema{Attributes: hybridDeploymentAgentAttribute().GetDatasourceSchema(),}
}

func HybridDeploymentAgentsDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema{
        Attributes: map[string]datasourceSchema.Attribute{
            "items": datasourceSchema.SetNestedAttribute{
                Computed: true,
                NestedObject: datasourceSchema.NestedAttributeObject{
                    Attributes: hybridDeploymentAgentAttribute().GetDatasourceSchema(),
                },
            },
        },
    }
}
