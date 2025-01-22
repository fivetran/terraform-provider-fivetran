package schema

import (
	"context"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TransformationProjectResource(ctx context.Context) resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: transformationProjectSchema().GetResourceSchema(),
		Blocks:     map[string]resourceSchema.Block{
			"project_config": resourceSchema.SingleNestedBlock{
				Attributes: transformationProjectConfigSchema().GetResourceSchema(),
			},
		},
	}
}

func TransformationProjectDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: transformationProjectSchema().GetDatasourceSchema(),
		Blocks:     map[string]datasourceSchema.Block{
			"project_config": datasourceSchema.SingleNestedBlock{
				Attributes: transformationProjectConfigSchema().GetDatasourceSchema(),
			},
		},
	}
}

func TransformationProjectListDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"projects": datasourceSchema.ListNestedAttribute{
				Computed: true,
				NestedObject: datasourceSchema.NestedAttributeObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the transformation project within the Fivetran system.",
						},
						"group_id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The name of the group within your account related to the project.",
						},
						"created_at": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The timestamp of when the project was created in your account.",
						},
						"created_by_id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the User within the Fivetran system who created the transformation Project.",
						},
						"type": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "Transformation project type.",
						},
					},
				},
			},
		},
	}
}

func transformationProjectSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the transformation Project within the Fivetran system.",
			},
			"group_id": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The unique identifier for the group within the Fivetran system.",
			},
			"type": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "Transformation project type.",
			},
			"status": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "Status of transformation Project (NOT_READY, READY, ERROR).",
			},
			"created_at": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The timestamp of the transformation Project creation.",
			},
			"created_by_id": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "The unique identifier for the User within the Fivetran system who created the dbt Project.",
			},
			"errors": {
				ValueType:   core.StringsSet,
				Readonly:    true,
				Description: "List of environment variables defined as key-value pairs in the raw string format using = as a separator. The variable name should have the DBT_ prefix and can contain A-Z, 0-9, dash, underscore, or dot characters. Example: \"DBT_VARIABLE=variable_value\"",
			},
			"run_tests": {
				ValueType:   core.Boolean,
				ResourceOnly:true,
				Description: "Specifies whether the setup tests should be run automatically. The default value is TRUE.",
			},
		},
	}
}

func transformationProjectConfigSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"dbt_version": {
				ValueType:   core.String,
				ForceNew:    true,
				Description: "The version of transformation that should run the project",
			},
			"default_schema": {
				ValueType:   core.String,
				ForceNew:    true,
				Description: "Default schema in destination. This production schema will contain your transformed data.",
			},
			"git_remote_url": {
				ValueType:   core.String,
				ForceNew:    true,
				Description: "Git remote URL with your transformation project",
			},
			"folder_path": {
				ValueType: core.String, 
				Description: "Folder in Git repo with your transformation project",
			},
			"git_branch":  {
				ValueType: core.String, 
				Description: "Git branch",
			},
			"threads": {
				ValueType:   core.Integer,
				Description: "The number of threads transformation will use (from 1 to 32). Make sure this value is compatible with your destination type. For example, Snowflake supports only 8 concurrent queries on an X-Small warehouse.",
			},
			"target_name": {
				ValueType:   core.String,
				Description: "Target name to set or override the value from the deployment.yaml",
			},
			"environment_vars": {
				ValueType:   core.StringsSet,
				Description: "List of environment variables defined as key-value pairs in the raw string format using = as a separator. The variable name should have the DBT_ prefix and can contain A-Z, 0-9, dash, underscore, or dot characters. Example: \"DBT_VARIABLE=variable_value\"",
			},
			"public_key": {
				ValueType:   core.String,
				Readonly:    true,
				Description: "Public key to grant Fivetran SSH access to git repository.",
			},

		},
	}
}