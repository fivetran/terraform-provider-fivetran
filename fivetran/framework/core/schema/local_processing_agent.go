package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func localDataProcessingAgentAttributesSchema() core.Schema {
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
			"re_auth": {
				ResourceOnly: true,
				ValueType:    core.Boolean,
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

func localDataProcessingAgentUsageSchema() core.Schema {
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

func LocalProcessingAgentResource() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: localDataProcessingAgentAttributesSchema().GetResourceSchema(),
		Blocks: map[string]resourceSchema.Block{
			"usage": resourceSchema.SetNestedBlock{
				NestedObject: resourceSchema.NestedBlockObject{
					Attributes: localDataProcessingAgentUsageSchema().GetResourceSchema(),
				},
			},
		},
	}
}

func LocalProcessingAgentDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: localDataProcessingAgentAttributesSchema().GetDatasourceSchema(),
		Blocks: map[string]datasourceSchema.Block{
			"usage": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: localDataProcessingAgentUsageSchema().GetDatasourceSchema(),
				},
			},
		},
	}
}

func LocalProcessingAgentsDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Blocks: map[string]datasourceSchema.Block{
			"items": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: localDataProcessingAgentAttributesSchema().GetDatasourceSchema(),
					Blocks: map[string]datasourceSchema.Block{
						"usage": datasourceSchema.SetNestedBlock{
							NestedObject: datasourceSchema.NestedBlockObject{
								Attributes: localDataProcessingAgentUsageSchema().GetDatasourceSchema(),
							},
						},
					},
				},
			},
		},
	}
}
