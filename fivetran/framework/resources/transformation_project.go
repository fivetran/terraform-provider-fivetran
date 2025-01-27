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
	resp.Schema = fivetranSchema.TransformationProjectResource(ctx)
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

	var data model.TransformationResourceProject
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
		resp.Diagnostics.AddError(
			"Unable to Create Transformation Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)
		return			
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

	var data model.TransformationResourceProject

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

	var state model.TransformationResourceProject
	var plan model.TransformationResourceProject

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewTransformationProjectUpdate()
	svc.ProjectId(state.Id.ValueString())
	
	runTestsPlan := core.GetBoolOrDefault(plan.RunTests, true)
	runTestsState := core.GetBoolOrDefault(state.RunTests, true)

	if runTestsPlan != runTestsState {
		svc.RunTests(runTestsPlan)
	}

	if !plan.ProjectConfig.IsUnknown() && !state.ProjectConfig.Equal(plan.ProjectConfig) {
		hasChanges := false
		projectConfig := fivetran.NewTransformationProjectConfig()
		configPlanAttributes := plan.ProjectConfig.Attributes()
		configStateAttributes := state.ProjectConfig.Attributes()

		fmt.Printf("configPlanAttributes %v\n", configPlanAttributes)
		fmt.Printf("configStateAttributes %v\n", configStateAttributes)
		if !configPlanAttributes["folder_path"].IsNull() &&
		!configPlanAttributes["folder_path"].IsUnknown() && 
		!configStateAttributes["folder_path"].(basetypes.StringValue).Equal(configPlanAttributes["folder_path"].(basetypes.StringValue)) {
			hasChanges = true
			projectConfig.FolderPath(configPlanAttributes["folder_path"].(basetypes.StringValue).ValueString())	
		}

		if !configPlanAttributes["git_branch"].IsNull() &&
		!configPlanAttributes["git_branch"].IsUnknown() && 
		!configStateAttributes["git_branch"].(basetypes.StringValue).Equal(configPlanAttributes["git_branch"].(basetypes.StringValue)) {
			hasChanges = true
			projectConfig.GitBranch(configPlanAttributes["git_branch"].(basetypes.StringValue).ValueString())	
		}

		if !configPlanAttributes["target_name"].IsNull() &&
		!configPlanAttributes["target_name"].IsUnknown() && 
		!configStateAttributes["target_name"].(basetypes.StringValue).Equal(configPlanAttributes["target_name"].(basetypes.StringValue)) {
			hasChanges = true
			projectConfig.TargetName(configPlanAttributes["target_name"].(basetypes.StringValue).ValueString())	
		}

		if !configPlanAttributes["threads"].IsNull() &&
		!configPlanAttributes["threads"].IsUnknown() && 
		!configStateAttributes["threads"].(basetypes.Int64Value).Equal(configPlanAttributes["threads"].(basetypes.Int64Value)) {
			hasChanges = true
			projectConfig.Threads(int(configPlanAttributes["threads"].(basetypes.Int64Value).ValueInt64()))
		}

		if !configPlanAttributes["environment_vars"].IsNull() &&
		!configPlanAttributes["environment_vars"].IsUnknown() && 
		!configStateAttributes["environment_vars"].(basetypes.SetValue).Equal(configPlanAttributes["environment_vars"].(basetypes.SetValue)) {
			evars := []string{}
			for _, ev := range configPlanAttributes["environment_vars"].(basetypes.SetValue).Elements() {
				evars = append(evars, ev.(basetypes.StringValue).ValueString())
			}
			hasChanges = true
			projectConfig.EnvironmentVars(evars)
		}

		if hasChanges {
			svc.ProjectConfig(projectConfig)			
		}
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

	var data model.TransformationResourceProject

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
