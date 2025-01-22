package resources

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TransformationProject() resource.Resource {
	return &transformationProject{}
}

type transformationProject struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &transformationProject{}
var _ resource.ResourceWithImportState = &transformationProject{}

func (r *transformationProject) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "fivetran_transformation_project"
}

func (r *transformationProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.DbtProjectResource(ctx)
}

func (r *transformationProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *transformationProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TransformationProject
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client := r.GetClient()
	svc := client.NewTransformationProjectCreate()

	svc.GroupId(data.GroupId.ValueString())
	svc.ProjectType(data.Type.ValueString())
	svc.RunTests(data.RunTests.ValueBool())

	if !data.ProjectConfig.IsNull() && !data.ProjectConfig.IsUnknown() {
		projectConfig := fivetran.NewTransformationProjectConfig()
		projectConfigAttributes := data.ProjectConfig.Attributes()
		projectConfig.DbtVersion(projectConfigAttributes["dbt_version"].(basetypes.StringValue).ValueString())
		projectConfig.DefaultSchema(projectConfigAttributes["default_schema"].(basetypes.StringValue).ValueString())
		projectConfig.GitRemoteUrl(projectConfigAttributes["git_remote_url"].(basetypes.StringValue).ValueString())
		projectConfig.FolderPath(projectConfigAttributes["folder_path"].(basetypes.StringValue).ValueString())
		projectConfig.GitBranch(projectConfigAttributes["git_branch"].(basetypes.StringValue).ValueString())
		projectConfig.TargetName(projectConfigAttributes["target_name"].(basetypes.StringValue).ValueString())
		projectConfig.Threads(int(projectConfigAttributes["threads"].(basetypes.Int64Value).ValueInt64()))

		if !projectConfigAttributes["environment_vars"].IsUnknown() && !projectConfigAttributes["environment_vars"].IsNull() {
			evars := []string{}
			for _, ev := range projectConfigAttributes["environment_vars"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			projectConfig.EnvironmentVars(evars)
		}

		svc.ProjectConfig(projectConfig)
	}

	projectResponse, err := svc.Do(ctx)
	if err != nil {
		if projectResponse.Code != "DbtProjectExists" {
			resp.Diagnostics.AddError(
				"Unable to Create dbt Project Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
			)

			return			
		} else {
			// try to recover Id
			projectListResponse, err := r.GetClient().NewTransformationProjectsList().Do(ctx)

			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Read Transformation Project Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
				)
				return
			}

			for _, v := range projectListResponse.Data.Items {
				if v.GroupId == data.GroupId.ValueString() {
					projectResponse, err := r.GetClient().NewTransformationProjectDetails().ProjectId(v.Id).Do(ctx)

					if err != nil {
						resp.Diagnostics.AddError(
							"Unable to Read Transformation Project Resource.",
							fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
						)
						return
					}
				}
			}
		}
	}

	data.ReadFromResponse(ctx, projectResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		// Do cleanup on error
		deleteResponse, err := client.NewTransformationProjectDelete().ProjectId(projectResponse.Data.Id).Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Cleanup Transformation Project Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
			)
		}
	}
}

func (r *transformationProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TransformationProject

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectResponse, err := r.GetClient().NewTransformationProjectDetails().ProjectId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Transformation Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, projectResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *transformationProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var state model.TransformationProject
	var plan model.TransformationProject

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewTransformationProjectUpdate()
	svc.ProjectId(state.Id.ValueString())
	svc.RunTests(plan.RunTests.ValueBool())

	if !plan.ProjectConfig.IsUnknown() && !state.ProjectConfig.Equal(plan.ProjectConfig) {
		projectConfig := fivetran.NewTransformationProjectConfig()
		projectConfigAttributes := plan.ProjectConfig.Attributes()
		projectConfig.FolderPath(projectConfigAttributes["folder_path"].(basetypes.StringValue).ValueString())
		projectConfig.GitBranch(projectConfigAttributes["git_branch"].(basetypes.StringValue).ValueString())
		projectConfig.TargetName(projectConfigAttributes["target_name"].(basetypes.StringValue).ValueString())
		projectConfig.Threads(int(projectConfigAttributes["threads"].(basetypes.Int64Value).ValueInt64()))

		if !projectConfigAttributes["environment_vars"].IsUnknown() && !projectConfigAttributes["environment_vars"].IsNull() {
			evars := []string{}
			for _, ev := range projectConfigAttributes["environment_vars"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			projectConfig.EnvironmentVars(evars)
		}
		svc.ProjectConfig(projectConfig)
	}

	projectResponse, err := svc.Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Transformation Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)
		return
	}

	plan.ReadFromResponse(ctx, projectResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *transformationProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.TransformationProject

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deleteResponse, err := r.GetClient().NewTransformationProjectDelete().ProjectId(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete transformation Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}
