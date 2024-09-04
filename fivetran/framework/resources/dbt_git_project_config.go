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
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DbtGitProjectConfig() resource.Resource {
	return &dbtGitProjectConfig{}
}

type dbtGitProjectConfig struct {
	core.ProviderResource
}

// Ensure the implementation satisfies the desired interfaces.
var _ resource.ResourceWithConfigure = &dbtGitProjectConfig{}
var _ resource.ResourceWithImportState = &dbtGitProjectConfig{}

func (r *dbtGitProjectConfig) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "fivetran_dbt_git_project_config"
}

func (r *dbtGitProjectConfig) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fivetranSchema.DbtGitProjectConfigSchema()
}

func (r *dbtGitProjectConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dbtGitProjectConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}
	client := r.GetClient()

	var data model.DbtGitProjectConfig
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewDbtProjectModify().DbtProjectID(data.ProjectId.ValueString())
	projectConfig := fivetran.NewDbtProjectConfig()
	projectConfig.GitRemoteUrl(data.GitRemoteUrl.ValueString())
	projectConfig.FolderPath(data.FolderPath.ValueString())
	projectConfig.GitBranch(data.GitBranch.ValueString())
	svc.ProjectConfig(projectConfig)

	projectResponse, err := svc.Do(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create dbt Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)

		return
	}

	data.ReadFromResponse(projectResponse)

	if data.EnsureReadiness.Equal(types.BoolValue(false)) {
		if ok, _ := ensureProjectIsReady(ctx, client, data.ProjectId.ValueString()); !ok {
			resp.Diagnostics.AddWarning(
				"Unable to finish set up dbt Git Project Config Resource.",
				"Project not ready.",
			)
		} else {
			data.EnsureReadiness = types.BoolValue(true)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dbtGitProjectConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}

	var data model.DbtGitProjectConfig

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectResponse, err := r.GetClient().NewDbtProjectDetails().DbtProjectID(data.ProjectId.ValueString()).Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dbt Project Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)
		return
	}

	data.ReadFromResponse(projectResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *dbtGitProjectConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.GetClient() == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Fivetran Client",
			"Please report this issue to the provider developers.",
		)

		return
	}
	client := r.GetClient()

	var state model.DbtGitProjectConfig
	var plan model.DbtGitProjectConfig

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	svc := r.GetClient().NewDbtProjectModify().DbtProjectID(state.ProjectId.ValueString())

	if !state.GitRemoteUrl.Equal(plan.GitRemoteUrl) || 
	   !state.GitBranch.Equal(plan.GitBranch) || 
	   !state.FolderPath.Equal(plan.FolderPath) {
		projectConfig := fivetran.NewDbtProjectConfig()
		projectConfig.GitRemoteUrl(plan.GitRemoteUrl.ValueString())
		projectConfig.FolderPath(plan.FolderPath.ValueString())
		projectConfig.GitBranch(plan.GitBranch.ValueString())
		svc.ProjectConfig(projectConfig)
	}

	projectResponse, err := svc.Do(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Dbt Git Project Config Resource.",
			fmt.Sprintf("%v; code: %v; message: %v", err, projectResponse.Code, projectResponse.Message),
		)
		return
	}

	plan.ReadFromResponse(projectResponse)

	if plan.EnsureReadiness.Equal(types.BoolValue(false)) {
		if ok, _ := ensureProjectIsReady(ctx, client, plan.ProjectId.ValueString()); !ok {
			resp.Diagnostics.AddWarning(
				"Unable to finish set up dbt Git Project Config Resource.",
				"Project not ready.",
			)
		} else {
			plan.EnsureReadiness = types.BoolValue(true)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dbtGitProjectConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// no op
}

func ensureProjectIsReady(ctx context.Context, client *fivetran.Client, projectId string) (bool, string) {
	for {
		s, projectErrors, e := pollProjectStatus(ctx, client, projectId)
		if e != nil {
			return false, fmt.Sprintf("unable to get status for dbt project: %v error: %v", projectId, e)
		}
		if s != "not_ready" {
			if s != "ready" {
				return false, fmt.Sprintf("dbt project: %v has \"ERROR\" status after creation; errors: %v;", projectId, projectErrors)
			}
			break
		}
		if dl, ok := ctx.Deadline(); ok && time.Now().After(dl.Add(-20*time.Second)) {
			// deadline will be exceeded on next iteration - it's time to cleanup
			return false, fmt.Sprintf("project %v is stuck in \"NOT_READY\" status", projectId)

		}
		helpers.ContextDelay(ctx, 10*time.Second)
	}
	return true, ""
}

func pollProjectStatus(ctx context.Context, client *fivetran.Client, projectId string) (string, []string, error) {
	resp, err := client.NewDbtProjectDetails().DbtProjectID(projectId).Do(ctx)
	if err != nil {
		return "", []string{}, err
	}
	return strings.ToLower(resp.Data.Status), resp.Data.Errors, err
}
