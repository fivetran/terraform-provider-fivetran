package model

import (
    "context"

    "github.com/fivetran/go-fivetran/transformations"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type TransformationResourceProject struct {
    Id              types.String `tfsdk:"id"`
    GroupId         types.String `tfsdk:"group_id"`
    Type            types.String `tfsdk:"type"`
    Status          types.String `tfsdk:"status"`
    CreatedAt       types.String `tfsdk:"created_at"`
    CreatedById     types.String `tfsdk:"created_by_id"`
    Errors          types.Set    `tfsdk:"errors"`
    RunTests        types.Bool   `tfsdk:"run_tests"`
    ProjectConfig   types.Object `tfsdk:"project_config"`
}

type TransformationDatasourceProject struct {
    Id              types.String `tfsdk:"id"`
    GroupId         types.String `tfsdk:"group_id"`
    Type            types.String `tfsdk:"type"`
    Status          types.String `tfsdk:"status"`
    CreatedAt       types.String `tfsdk:"created_at"`
    CreatedById     types.String `tfsdk:"created_by_id"`
    Errors          types.Set    `tfsdk:"errors"`
    ProjectConfig   types.Object `tfsdk:"project_config"`
}

func (d *TransformationResourceProject) ReadFromResponse(ctx context.Context, resp transformations.TransformationProjectResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.GroupId = types.StringValue(resp.Data.GroupId)
    d.Type = types.StringValue(resp.Data.ProjectType)
    d.CreatedAt = types.StringValue(resp.Data.CreatedAt)
    d.CreatedById = types.StringValue(resp.Data.CreatedById)
    d.Status = types.StringValue(resp.Data.Status)
    
    errors := []attr.Value{}
    for _, el := range resp.Data.Errors {
        errors = append(errors, types.StringValue(el))
    }
    if len(errors) > 0 {
        d.Errors = types.SetValueMust(types.StringType, errors)
    } else {
        if d.Errors.IsUnknown() {
            d.Errors = types.SetNull(types.StringType)
        }
    }

    projectConfigTypes := map[string]attr.Type{
        "dbt_version":          types.StringType,
        "default_schema":       types.StringType,
        "git_remote_url":       types.StringType,
        "folder_path":          types.StringType,
        "git_branch":           types.StringType,
        "target_name":          types.StringType,
        "environment_vars":     types.SetType{ElemType: types.StringType},
        "public_key":           types.StringType,
        "threads":              types.Int64Type,
    }
    projectConfigItems := map[string]attr.Value{}
    projectConfigItems["dbt_version"] = types.StringValue(resp.Data.ProjectConfig.DbtVersion)
    projectConfigItems["default_schema"] = types.StringValue(resp.Data.ProjectConfig.DefaultSchema)
    projectConfigItems["git_remote_url"] = types.StringValue(resp.Data.ProjectConfig.GitRemoteUrl)
    projectConfigItems["folder_path"] = types.StringValue(resp.Data.ProjectConfig.FolderPath)
    projectConfigItems["git_branch"] = types.StringValue(resp.Data.ProjectConfig.GitBranch)
    projectConfigItems["target_name"] = types.StringValue(resp.Data.ProjectConfig.TargetName)
    projectConfigItems["public_key"] = types.StringValue(resp.Data.ProjectConfig.PublicKey)
    projectConfigItems["threads"] = types.Int64Value(int64(resp.Data.ProjectConfig.Threads))

    envVars := []attr.Value{}
    for _, el := range resp.Data.ProjectConfig.EnvironmentVars {
        envVars = append(envVars, types.StringValue(el))
    }
    if len(envVars) > 0 {
        projectConfigItems["environment_vars"] = types.SetValueMust(types.StringType, envVars)
    } else {
        projectConfigItems["environment_vars"] = types.SetNull(types.StringType)
    }
        
    d.ProjectConfig, _ = types.ObjectValue(projectConfigTypes, projectConfigItems)
}

func (d *TransformationDatasourceProject) ReadFromResponse(ctx context.Context, resp transformations.TransformationProjectResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.GroupId = types.StringValue(resp.Data.GroupId)
    d.Type = types.StringValue(resp.Data.ProjectType)
    d.CreatedAt = types.StringValue(resp.Data.CreatedAt)
    d.CreatedById = types.StringValue(resp.Data.CreatedById)
    d.Status = types.StringValue(resp.Data.Status)
    
    errors := []attr.Value{}
    for _, el := range resp.Data.Errors {
        errors = append(errors, types.StringValue(el))
    }
    if len(errors) > 0 {
        d.Errors = types.SetValueMust(types.StringType, errors)
    } else {
        if d.Errors.IsUnknown() {
            d.Errors = types.SetNull(types.StringType)
        }
    }

    projectConfigTypes := map[string]attr.Type{
        "dbt_version":          types.StringType,
        "default_schema":       types.StringType,
        "git_remote_url":       types.StringType,
        "folder_path":          types.StringType,
        "git_branch":           types.StringType,
        "target_name":          types.StringType,
        "environment_vars":     types.SetType{ElemType: types.StringType},
        "public_key":           types.StringType,
        "threads":              types.Int64Type,
    }
    projectConfigItems := map[string]attr.Value{}
    projectConfigItems["dbt_version"] = types.StringValue(resp.Data.ProjectConfig.DbtVersion)
    projectConfigItems["default_schema"] = types.StringValue(resp.Data.ProjectConfig.DefaultSchema)
    projectConfigItems["git_remote_url"] = types.StringValue(resp.Data.ProjectConfig.GitRemoteUrl)
    projectConfigItems["folder_path"] = types.StringValue(resp.Data.ProjectConfig.FolderPath)
    projectConfigItems["git_branch"] = types.StringValue(resp.Data.ProjectConfig.GitBranch)
    projectConfigItems["target_name"] = types.StringValue(resp.Data.ProjectConfig.TargetName)
    projectConfigItems["public_key"] = types.StringValue(resp.Data.ProjectConfig.PublicKey)
    projectConfigItems["threads"] = types.Int64Value(int64(resp.Data.ProjectConfig.Threads))

    envVars := []attr.Value{}
    for _, el := range resp.Data.ProjectConfig.EnvironmentVars {
        envVars = append(envVars, types.StringValue(el))
    }
    if len(envVars) > 0 {
        projectConfigItems["environment_vars"] = types.SetValueMust(types.StringType, envVars)
    } else {
        projectConfigItems["environment_vars"] = types.SetNull(types.StringType)
    }
        
    d.ProjectConfig, _ = types.ObjectValue(projectConfigTypes, projectConfigItems)
}