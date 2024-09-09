package model

import (
	"github.com/fivetran/go-fivetran/dbt"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DbtGitProjectConfig struct {
	Id              types.String `tfsdk:"id"`
	ProjectId       types.String `tfsdk:"project_id"`
	GitRemoteUrl    types.String `tfsdk:"git_remote_url"`
	GitBranch       types.String `tfsdk:"git_branch"`
	FolderPath      types.String `tfsdk:"folder_path"`
	EnsureReadiness types.Bool   `tfsdk:"ensure_readiness"`
}

func (d *DbtGitProjectConfig) ReadFromResponse(resp dbt.DbtProjectDetailsResponse) {
	d.Id = types.StringValue(resp.Data.ID)
	d.ProjectId = types.StringValue(resp.Data.ID)

	if resp.Data.ProjectConfig.GitRemoteUrl != "" {
		d.GitRemoteUrl = types.StringValue(resp.Data.ProjectConfig.GitRemoteUrl)
	} else {
		d.GitRemoteUrl = types.StringNull()
	}

	if resp.Data.ProjectConfig.GitBranch != "" {
		d.GitBranch = types.StringValue(resp.Data.ProjectConfig.GitBranch)
	} else {
		d.GitBranch = types.StringNull()
	}

	if resp.Data.ProjectConfig.FolderPath != "" {
		d.FolderPath = types.StringValue(resp.Data.ProjectConfig.FolderPath)
	} else {
		d.FolderPath = types.StringNull()
	}

	if d.EnsureReadiness.IsUnknown() {
		d.EnsureReadiness = types.BoolValue(false)
	}
}
