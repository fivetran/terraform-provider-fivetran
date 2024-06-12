package schema

import (
    "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
    resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func ProxyAgentSchema() core.Schema {
    return core.Schema{
        Fields: map[string]core.SchemaField{
            "id": {
                IsId:        true,
                ValueType:   core.String,
                Description: "The unique identifier for the proxy within your account.",
            },
            "group_region":{
                Required:    true,
                ForceNew:    true,
                ValueType:   core.String,
                Description: "Data processing location. This is where Fivetran will operate and run computation on data.",
            },
            "display_name":{
                Required:    true,
                ForceNew:    true,
                ValueType:   core.String,
                Description: "Proxy agent name.",
            },
            "registred_at": {
                Readonly:    true,
                ValueType:   core.String,
                Description: "The timestamp of the time the proxy agent was created in your account.",
            },
            "created_by": {
                Readonly:    true,
                ValueType:   core.String,
                Description: "The actor who created the proxy agent.",
            },
            "account_id": {
                Readonly:    true,
                ValueType:   core.String,
                Description: "The unique identifier for the account.",
            },
            "proxy_server_uri": {
                Readonly:    true,
                ValueType:   core.String,
                ResourceOnly:true,
                Description: "The proxy server URI.",
            },
            "token": {
                Readonly:    true,
                ValueType:   core.String,
                Description: "The auth token.",
            },
            "salt": {
                Readonly:    true,
                ValueType:   core.String,
                Description: "The salt.",
            },
        },
    }
}

func ProxyAgentResource() resourceSchema.Schema {
    return resourceSchema.Schema{
        Attributes: ProxyAgentSchema().GetResourceSchema(),
    }
}

func ProxyAgentDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema{
        Attributes: ProxyAgentSchema().GetDatasourceSchema(),
    }
}

func ProxyAgentsDatasource() datasourceSchema.Schema {
    return datasourceSchema.Schema{
        Blocks: map[string]datasourceSchema.Block{
            "items": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: ProxyAgentSchema().GetDatasourceSchema(),
                },
            },
        },
    }
}