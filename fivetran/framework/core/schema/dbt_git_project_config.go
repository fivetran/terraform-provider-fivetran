package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func DbtGitProjectConfigSchema() resourceSchema.Schema {
	return resourceSchema.Schema{
		Attributes: dbtGitProjectConfigSchema().GetResourceSchema(),
	}
}

func dbtGitProjectConfigSchema() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the dbt Project within the Fivetran system.",
			},
			"project_id": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The unique identifier for the dbt Project within the Fivetran system.",
			},
			"git_remote_url": {
				ValueType:   core.String,
				Description: "Git remote URL with your dbt project.",
			},
			"git_branch":  {
				ValueType: core.String,
				Description: "Git branch.",
			},
			"folder_path": {
				ValueType: core.String, 
				Description: "Folder in Git repo with your dbt project.",
			},
			"ensure_readiness": {
				ValueType:   core.Boolean,
				Description: "Should resource wait for project to finish initialization. Default value: false.",
			},
		},
	}
}
