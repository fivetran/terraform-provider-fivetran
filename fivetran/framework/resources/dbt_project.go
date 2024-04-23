package resources

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/model"
	fivetranSchema "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/schema"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/datasources"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func DbtProject() resource.Resource {
	return &dbtProject{}
}

type dbtProject struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &dbtProject{}
var _ resource.ResourceWithImportState = &dbtProject{}

func (r *dbtProject) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "fivetran_dbt_project"
}

func (r *dbtProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.DbtProjectResource(ctx)
}

func (r *dbtProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dbtProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtProjectResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client := r.GetClient()
	svc := client.NewDbtProjectCreate()

	svc.GroupID(data.GroupId.ValueString())
	svc.DbtVersion(data.DbtVersion.ValueString())
	svc.DefaultSchema(data.DefaultSchema.ValueString())

	projectConfigAttributes := data.ProjectConfig.Attributes()

	gitRemoteUrl := projectConfigAttributes["git_remote_url"]
	gitRemoteUrlDefined := !gitRemoteUrl.IsUnknown() && !gitRemoteUrl.IsNull()

	// If project type not defined we consider project_type = "GIT" on API side
	projectType := "GIT"
	if !data.Type.IsUnknown() && !data.Type.IsNull() {
		projectType = data.Type.ValueString()
	}

	if projectType != "GIT" {
		resp.Diagnostics.AddError(
			"Unable to Create dbt Project.",
			"Only `GIT` project type supported.",
		)
		return
	}

	svc.Type(projectType)
	// Currently git_remote_url is required: only GIT project could be managed via API
	if !gitRemoteUrlDefined && projectType == "GIT" {
		resp.Diagnostics.AddError(
			"Unable to Create dbt Project.",
			"Field `git_remote_url` is required for project of type GIT.",
		)
		return
	}
	projectConfig := fivetran.NewDbtProjectConfig().
		GitRemoteUrl(gitRemoteUrl.(basetypes.StringValue).ValueString())

	if v, ok := projectConfigAttributes["git_branch"].(basetypes.StringValue); ok && !v.IsUnknown() && !v.IsNull() {
		projectConfig.GitBranch(v.ValueString())
	}

	if v, ok := projectConfigAttributes["folder_path"].(basetypes.StringValue); ok && !v.IsUnknown() && !v.IsNull() {
		projectConfig.FolderPath(v.ValueString())
	}
	svc.ProjectConfig(projectConfig)

	if !data.EnvironmentVars.IsUnknown() && !data.EnvironmentVars.IsNull() {
		evars := []string{}
		for _, ev := range data.EnvironmentVars.Elements() {
			evars = append(evars, ev.(basetypes.StringValue).ValueString())
		}
		svc.EnvironmentVars(evars)
	}

	if !data.TargetName.IsNull() && !data.TargetName.IsUnknown() {
		svc.TargetName(data.TargetName.ValueString())
	}

	if !data.Threads.IsNull() && !data.Threads.IsUnknown() {
		svc.Threads(int(data.Threads.ValueInt64()))
	}

	projectResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create dbt Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)

		return
	}

	data.ReadFromResponse(ctx, projectResponse, nil)

	projectIsReady := strings.ToLower(projectResponse.Data.Status) == "ready"

	if data.EnsureReadiness.Equal(types.BoolValue(true)) && strings.ToLower(projectResponse.Data.Status) != "ready" {
		if ok := ensureProjectIsReady(resp, ctx, client, projectResponse.Data.ID); !ok {
			resp.Diagnostics.AddError(
				"Unable to set up dbt Project Resource.",
				"Project not ready.",
			)
			deleteResponse, err := client.NewDbtProjectDelete().DbtProjectID(projectResponse.Data.ID).Do(ctx)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to Cleanup dbt Project Resource.",
					fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
				)
			}
			return
		} else {
			projectIsReady = true
		}
	}

	if projectIsReady {
		modelsResp, err := datasources.GetAllDbtModelsForProject(r.GetClient(), ctx, projectResponse.Data.ID, 1000)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"DbtProject Models Read Error.",
				fmt.Sprintf("%v; code: %v; message: %v", err, modelsResp.Code, modelsResp.Message),
			)
		} else {
			data.ReadFromResponse(ctx, projectResponse, &modelsResp)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		// Do cleanup on error
		deleteResponse, err := client.NewDbtProjectDelete().DbtProjectID(projectResponse.Data.ID).Do(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Cleanup dbt Project Resource.",
				fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
			)
		}
	}
}

func (r *dbtProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectResponse, err := r.GetClient().NewDbtProjectDetails().DbtProjectID(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dbt Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)
		return
	}

	data.ReadFromResponse(ctx, projectResponse, nil)

	if strings.ToLower(projectResponse.Data.Status) == "ready" {
		modelsResp, err := datasources.GetAllDbtModelsForProject(r.GetClient(), ctx, projectResponse.Data.ID, 1000)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"DbtProject Models Read Error.",
				fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
			)
		} else {
			data.ReadFromResponse(ctx, projectResponse, &modelsResp)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dbtProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var state model.DbtProjectResourceModel
	var plan model.DbtProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewDbtProjectModify().DbtProjectID(state.Id.ValueString())

	if !state.DbtVersion.Equal(plan.DbtVersion) {
		svc.DbtVersion(plan.DbtVersion.ValueString())
	}

	if !state.TargetName.Equal(plan.TargetName) {
		svc.TargetName(plan.TargetName.ValueString())
	}

	if !state.Threads.Equal(plan.Threads) {
		svc.Threads(int(plan.Threads.ValueInt64()))
	}

	if !state.EnvironmentVars.Equal(plan.EnvironmentVars) {
		evars := []string{}
		for _, ev := range plan.EnvironmentVars.Elements() {
			evars = append(evars, ev.(basetypes.StringValue).ValueString())
		}
		svc.EnvironmentVars(evars)
	}

	if !state.ProjectConfig.Equal(plan.ProjectConfig) {
		planConfigAttributes := plan.ProjectConfig.Attributes()
		projectConfig := fivetran.NewDbtProjectConfig()
		projectConfig.FolderPath(planConfigAttributes["folder_path"].(basetypes.StringValue).ValueString())
		projectConfig.GitBranch(planConfigAttributes["git_branch"].(basetypes.StringValue).ValueString())
		svc.ProjectConfig(projectConfig)
	}

	projectResponse, err := svc.Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Dbt Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)
		return
	}

	plan.ReadFromResponse(ctx, projectResponse, nil)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

func (r *dbtProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtProjectResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deleteResponse, err := r.GetClient().NewDbtProjectDelete().DbtProjectID(data.Id.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete dbt Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, deleteResponse.Code, deleteResponse.Message),
		)
		return
	}
}

func ensureProjectIsReady(
	resp *resource.CreateResponse,
	ctx context.Context,
	client *fivetran.Client,
	projectId string) bool {
	for {
		s, projectErrors, e := pollProjectStatus(ctx, client, projectId)
		if e != nil {

			resp.Diagnostics.AddError("create error", fmt.Sprintf("unable to get status for dbt project: %v error: %v", projectId, e))
			return false

		}
		if s != "not_ready" {
			if s != "ready" {

				resp.Diagnostics.AddError("create error", fmt.Sprintf("dbt project: %v has \"ERROR\" status after creation; errors: %v;", projectId, projectErrors))
				return false

			}
			break
		}
		if dl, ok := ctx.Deadline(); ok && time.Now().After(dl.Add(-20*time.Second)) {
			// deadline will be exceeded on next iteration - it's time to cleanup

			resp.Diagnostics.AddError("create error", fmt.Sprintf("project %v is stuck in \"NOT_READY\" status", projectId))
			return false

		}
		helpers.ContextDelay(ctx, 10*time.Second)
	}
	return true
}

func pollProjectStatus(ctx context.Context, client *fivetran.Client, projectId string) (string, []string, error) {
	resp, err := client.NewDbtProjectDetails().DbtProjectID(projectId).Do(ctx)
	if err != nil {
		return "", []string{}, err
	}
	return strings.ToLower(resp.Data.Status), resp.Data.Errors, err
}
