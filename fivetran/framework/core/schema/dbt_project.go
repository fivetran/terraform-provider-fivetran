package schema

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func DbtProjectResource(ctx context.Context) resourceSchema.Schema {
	attributes := dbtProjectSchema().GetResourceSchema()
	attributes["models"] = resourceSchema.SetNestedAttribute{
		Computed: true,
		NestedObject: resourceSchema.NestedAttributeObject{
			Attributes: DbtModelSchema().GetResourceSchema(),
		},
	}
	return resourceSchema.Schema{
		Attributes: attributes,
		Blocks:     dbtProjectResourceBlocks(ctx),
		Version:    1,
	}
}

func DbtProjectDatasource() datasourceSchema.Schema {
	attributes := dbtProjectSchema().GetDatasourceSchema()
	attributes["models"] = datasourceSchema.SetNestedAttribute{
		Computed: true,
		NestedObject: datasourceSchema.NestedAttributeObject{
			Attributes: DbtModelSchema().GetDatasourceSchema(),
		},
	}
	return datasourceSchema.Schema{
		Attributes: attributes,
		Blocks:     dbtProjectDatasourceBlocks(),
	}
}

func dbtProjectSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the dbt Project within the Fivetran system.",
			},
			"group_id": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The unique identifier for the group within the Fivetran system.",
			},
			"default_schema": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "Default schema in destination. This production schema will contain your transformed data.",
			},
			"dbt_version": {
				Required:    true,
				ValueType:   core.String,
				Description: "The version of dbt that should run the project. We support the following versions: 0.18.0 - 0.18.2, 0.19.0 - 0.19.2, 0.20.0 - 0.20.2, 0.21.0 - 0.21.1, 1.0.0, 1.0.1, 1.0.3 - 1.0.9, 1.1.0 - 1.1.3, 1.2.0 - 1.2.4, 1.3.0 - 1.3.2, 1.4.1.",
			},

			"environment_vars": {
				ValueType:   core.StringsSet,
				Description: "List of environment variables defined as key-value pairs in the raw string format using = as a separator. The variable name should have the DBT_ prefix and can contain A-Z, 0-9, dash, underscore, or dot characters. Example: \"DBT_VARIABLE=variable_value\"",
			},
			"target_name": {
				ValueType:   core.String,
				Description: "Target name to set or override the value from the deployment.yaml",
			},
			"threads": {
				ValueType:   core.Integer,
				Description: "The number of threads dbt will use (from 1 to 32). Make sure this value is compatible with your destination type. For example, Snowflake supports only 8 concurrent queries on an X-Small warehouse.",
			},
			"type": {
				ValueType:   core.String,
				ForceNew:    true,
				Description: "Type of dbt Project. Currently only `GIT` supported. Empty value will be considered as default (GIT).",
			},
			"status": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "Status of dbt Project (NOT_READY, READY, ERROR).",
			},
			"created_at": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The timestamp of the dbt Project creation.",
			},
			"created_by_id": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The unique identifier for the User within the Fivetran system who created the dbt Project.",
			},
			"public_key": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "Public key to grant Fivetran SSH access to git repository.",
			},
			"ensure_readiness": {
				ValueType:   core.Boolean,
				Description: "Should resource wait for project to finish initialization. Default value: true.",
			},
		},
	}
}

func dbtProjectConfigSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"git_remote_url": {
				ValueType:   core.String,
				ForceNew:    true, // git_remote_url can't be changed after project creation
				Description: "Git remote URL with your dbt project."},
			"git_branch":  {ValueType: core.String, Description: "Git branch."},
			"folder_path": {ValueType: core.String, Description: "Folder in Git repo with your dbt project."},
		},
	}
}

func dbtProjectResourceSchema() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: dbtProjectSchema().GetResourceSchema(),
	}
}

func dbtProjectResourceBlocks(ctx context.Context) map[string]resourceSchema.Block {
	return map[string]resourceSchema.Block{
		"project_config": resourceSchema.SingleNestedBlock{
			Attributes: dbtProjectConfigSchema().GetResourceSchema(),
		},
		"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true}),
	}
}

func dbtProjectDatasourceBlocks() map[string]datasourceSchema.Block {
	return map[string]datasourceSchema.Block{
		"project_config": datasourceSchema.SingleNestedBlock{
			Attributes: dbtProjectConfigSchema().GetDatasourceSchema(),
		},
	}
}
