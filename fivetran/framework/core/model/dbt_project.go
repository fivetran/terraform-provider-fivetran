package model

import (
	"context"
	"strings"

	"github.com/fivetran/go-fivetran/dbt"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DbtProject struct {
	Id              types.String `tfsdk:"id"`
	GroupId         types.String `tfsdk:"group_id"`
	DefaultSchema   types.String `tfsdk:"default_schema"`
	DbtVersion      types.String `tfsdk:"dbt_version"`
	EnvironmentVars types.Set    `tfsdk:"environment_vars"`
	TargetName      types.String `tfsdk:"target_name"`
	Threads         types.Int64  `tfsdk:"threads"`
	Type            types.String `tfsdk:"type"`
	Status          types.String `tfsdk:"status"`
	CreatedAt       types.String `tfsdk:"created_at"`
	CreatedById     types.String `tfsdk:"created_by_id"`
	PublicKey       types.String `tfsdk:"public_key"`
	ProjectConfig   types.Object `tfsdk:"project_config"`
	Models          types.Set    `tfsdk:"models"`
	EnsureReadiness types.Bool   `tfsdk:"ensure_readiness"`
}

type DbtProjectResourceModel struct {
	Id              types.String   `tfsdk:"id"`
	GroupId         types.String   `tfsdk:"group_id"`
	DefaultSchema   types.String   `tfsdk:"default_schema"`
	DbtVersion      types.String   `tfsdk:"dbt_version"`
	EnvironmentVars types.Set      `tfsdk:"environment_vars"`
	TargetName      types.String   `tfsdk:"target_name"`
	Threads         types.Int64    `tfsdk:"threads"`
	Type            types.String   `tfsdk:"type"`
	Status          types.String   `tfsdk:"status"`
	CreatedAt       types.String   `tfsdk:"created_at"`
	CreatedById     types.String   `tfsdk:"created_by_id"`
	PublicKey       types.String   `tfsdk:"public_key"`
	ProjectConfig   types.Object   `tfsdk:"project_config"`
	Models          types.Set      `tfsdk:"models"`
	EnsureReadiness types.Bool     `tfsdk:"ensure_readiness"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

func (d *DbtProjectResourceModel) ReadFromResponse(ctx context.Context, resp dbt.DbtProjectDetailsResponse, modelsResp *dbt.DbtModelsListResponse) {
	d.Id = types.StringValue(resp.Data.ID)
	d.GroupId = types.StringValue(resp.Data.GroupId)
	d.DefaultSchema = types.StringValue(resp.Data.DefaultSchema)
	d.DbtVersion = types.StringValue(resp.Data.DbtVersion)

	elements := []attr.Value{}
	for _, envVar := range resp.Data.EnvironmentVars {
		elements = append(elements, types.StringValue(envVar))
	}
	if len(elements) > 0 {
		d.EnvironmentVars = types.SetValueMust(types.StringType, elements)
	} else {
		if d.EnvironmentVars.IsUnknown() {
			d.EnvironmentVars = types.SetNull(types.StringType)
		}
	}

	if resp.Data.TargetName != "" {
		d.TargetName = types.StringValue(resp.Data.TargetName)
	} else {
		d.TargetName = types.StringNull()
	}
	d.Threads = types.Int64Value(int64(resp.Data.Threads))
	d.Type = types.StringValue(resp.Data.Type)
	d.Status = types.StringValue(resp.Data.Status)
	d.CreatedAt = types.StringValue(resp.Data.CreatedAt)
	d.CreatedById = types.StringValue(resp.Data.CreatedById)
	d.PublicKey = types.StringValue(resp.Data.PublicKey)

	projectConfigTypes := map[string]attr.Type{
		"git_remote_url": types.StringType,
		"git_branch":     types.StringType,
		"folder_path":    types.StringType,
	}
	projectConfigItems := map[string]attr.Value{
		"git_remote_url": types.StringValue(resp.Data.ProjectConfig.GitRemoteUrl),
		// "git_branch":     types.StringValue(resp.Data.ProjectConfig.GitBranch),
		// "folder_path":    types.StringValue(resp.Data.ProjectConfig.FolderPath),
	}

	if resp.Data.ProjectConfig.GitBranch != "" {
		projectConfigItems["git_branch"] = types.StringValue(resp.Data.ProjectConfig.GitBranch)
	} else {
		projectConfigItems["git_branch"] = types.StringNull()
	}

	if resp.Data.ProjectConfig.FolderPath != "" {
		projectConfigItems["folder_path"] = types.StringValue(resp.Data.ProjectConfig.FolderPath)
	} else {
		projectConfigItems["folder_path"] = types.StringNull()
	}

	d.ProjectConfig, _ = types.ObjectValue(projectConfigTypes, projectConfigItems)
	d.EnsureReadiness = types.BoolValue(strings.ToLower(resp.Data.Status) == "ready")
	if modelsResp != nil {
		d.Models = GetModelsSetFromResponse(*modelsResp)
	} else {
		if d.Models.IsUnknown() {
			d.Models = types.SetNull(types.ObjectType{AttrTypes: ModelElementType()})
		}
	}
}

func (d *DbtProject) ReadFromResponse(ctx context.Context, resp dbt.DbtProjectDetailsResponse, modelsResp *dbt.DbtModelsListResponse) {
	d.Id = types.StringValue(resp.Data.ID)
	d.GroupId = types.StringValue(resp.Data.GroupId)
	d.DefaultSchema = types.StringValue(resp.Data.DefaultSchema)
	d.DbtVersion = types.StringValue(resp.Data.DbtVersion)

	elements := []attr.Value{}
	for _, envVar := range resp.Data.EnvironmentVars {
		elements = append(elements, types.StringValue(envVar))
	}
	d.EnvironmentVars, _ = types.SetValue(types.StringType, elements)

	d.TargetName = types.StringValue(resp.Data.TargetName)
	d.Threads = types.Int64Value(int64(resp.Data.Threads))
	d.Type = types.StringValue(resp.Data.Type)
	d.Status = types.StringValue(resp.Data.Status)
	d.CreatedAt = types.StringValue(resp.Data.CreatedAt)
	d.CreatedById = types.StringValue(resp.Data.CreatedById)
	d.PublicKey = types.StringValue(resp.Data.PublicKey)

	projectConfigTypes := map[string]attr.Type{
		"git_remote_url": types.StringType,
		"git_branch":     types.StringType,
		"folder_path":    types.StringType,
	}
	projectConfigItems := map[string]attr.Value{
		"git_remote_url": types.StringValue(resp.Data.ProjectConfig.GitRemoteUrl),
		"git_branch":     types.StringValue(resp.Data.ProjectConfig.GitBranch),
		"folder_path":    types.StringValue(resp.Data.ProjectConfig.FolderPath),
	}

	d.ProjectConfig, _ = types.ObjectValue(projectConfigTypes, projectConfigItems)
	d.EnsureReadiness = types.BoolValue(strings.ToLower(resp.Data.Status) == "ready")
	if modelsResp != nil {
		d.Models = GetModelsSetFromResponse(*modelsResp)
	} else {
		if d.Models.IsUnknown() {
			d.Models = types.SetNull(types.ObjectType{AttrTypes: ModelElementType()})
		}
	}
}