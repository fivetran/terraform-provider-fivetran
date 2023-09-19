package fivetran

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDbtProject() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDbtProjectCreate,
		ReadContext:          resourceDbtProjectRead,
		UpdateContext:        resourceDbtProjectUpdate,
		DeleteContext:        resourceDbtProjectDelete,
		Importer:             &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema:               getDbtProjectSchema(false),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},
	}
}

func getDbtProjectSchema(datasource bool) map[string]*schema.Schema {
	maxItems := 1
	if datasource {
		maxItems = 0
	}
	result := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    !datasource,
			Required:    datasource,
			Description: "The unique identifier for the dbt Project within the Fivetran system.",
		},

		// required immutable
		"group_id": {
			Type:        schema.TypeString,
			Required:    !datasource,
			ForceNew:    !datasource,
			Computed:    datasource,
			Description: "The unique identifier for the group within the Fivetran system.",
		},
		"default_schema": {
			Type:        schema.TypeString,
			Required:    !datasource,
			ForceNew:    !datasource,
			Computed:    datasource,
			Description: "Default schema in destination. This production schema will contain your transformed data.",
		},

		// required
		"dbt_version": {
			Type:        schema.TypeString,
			Required:    !datasource,
			Computed:    datasource,
			Description: "The version of dbt that should run the project. We support the following versions: 0.18.0 - 0.18.2, 0.19.0 - 0.19.2, 0.20.0 - 0.20.2, 0.21.0 - 0.21.1, 1.0.0, 1.0.1, 1.0.3 - 1.0.9, 1.1.0 - 1.1.3, 1.2.0 - 1.2.4, 1.3.0 - 1.3.2, 1.4.1.",
		},
		"project_config": {
			Type:        schema.TypeList,
			MaxItems:    maxItems,
			Required:    !datasource,
			Computed:    datasource,
			Description: "Type specific dbt Project configuration parameters.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"git_remote_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    !datasource,
						ForceNew:    true, // git_remote_url can't be changed after project creation
						Description: "Git remote URL with your dbt project."},
					"git_branch":  {Type: schema.TypeString, Computed: true, Optional: !datasource, Description: "Git branch."},
					"folder_path": {Type: schema.TypeString, Computed: true, Optional: !datasource, Description: "Folder in Git repo with your dbt project."},
				},
			},
		},

		// optional
		"environment_vars": {
			Type:     schema.TypeSet,
			Optional: !datasource,
			Computed: datasource,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"target_name": {
			Type:        schema.TypeString,
			Optional:    !datasource,
			Computed:    datasource,
			Description: "Target name to set or override the value from the deployment.yaml",
		},
		"threads": {
			Type:        schema.TypeInt,
			Optional:    !datasource,
			Computed:    datasource,
			Description: "The number of threads dbt will use (from 1 to 32). Make sure this value is compatible with your destination type. For example, Snowflake supports only 8 concurrent queries on an X-Small warehouse.",
		},

		// immutable
		"type": {
			Type:        schema.TypeString,
			Optional:    !datasource,
			Computed:    true,
			ForceNew:    true, // project type can't be changed
			Description: "Type of dbt Project. Currently only `GIT` supported. Empty value will be considered as default (GIT).",
		},

		// readonly fields
		"status":        {Type: schema.TypeString, Computed: true, Description: "Status of dbt Project (NOT_READY, READY, ERROR)."},
		"created_at":    {Type: schema.TypeString, Computed: true, Description: "The timestamp of the dbt Project creation."},
		"created_by_id": {Type: schema.TypeString, Computed: true, Description: "The unique identifier for the User within the Fivetran system who created the dbt Project."},
		"public_key":    {Type: schema.TypeString, Computed: true, Description: "Public key to grant Fivetran SSH access to git repository."},

		"models": dbtModelsSchema(),
	}

	if !datasource {
		result["ensure_readiness"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Should resource wait for project to finish initialization.",
		}
	}
	return result
}

func resourceDbtProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx, cancel := setContextTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	client := m.(*fivetran.Client)
	svc := client.NewDbtProjectCreate()

	svc.GroupID(d.Get("group_id").(string))
	svc.DbtVersion(d.Get("dbt_version").(string))
	svc.DefaultSchema(d.Get("default_schema").(string))

	gitRemoteUrl, gitRemoteUrlDefined := d.GetOk("project_config.0.git_remote_url")

	// If project type not defined we consider project_type = "GIT" on API side
	projectType, ok := d.GetOk("type")
	if !ok {
		projectType = "GIT"
	}

	if projectType != "GIT" {
		return newDiagAppend(
			diags, diag.Error, "create error",
			fmt.Sprintf("%v; code: %v; message: %v", "", "", "Able to create only a project of type GIT"))
	}

	svc.Type(projectType.(string))
	// Currently git_remote_url is required: only GIT project could be managed via API
	if !gitRemoteUrlDefined && projectType == "GIT" {
		return newDiagAppend(
			diags, diag.Error, "create error",
			fmt.Sprintf("%v; code: %v; message: %v", "", "", "git_remote_url is required for project of type GIT"))
	}
	projectConfig := fivetran.NewDbtProjectConfig().
		GitRemoteUrl(gitRemoteUrl.(string))

	if v, ok := d.GetOk("project_config.0.git_branch"); ok {
		projectConfig.GitBranch(v.(string))
	}

	if v, ok := d.GetOk("project_config.0.folder_path"); ok {
		projectConfig.FolderPath(v.(string))
	}
	svc.ProjectConfig(projectConfig)

	if v, ok := d.GetOk("environment_vars"); ok {
		svc.EnvironmentVars(xInterfaceStrXStr(v.(*schema.Set).List()))
	}
	if v, ok := d.GetOk("target_name"); ok {
		svc.TargetName(v.(string))
	}
	if v, ok := d.GetOk("threads"); ok {
		svc.Threads(v.(int))
	}

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if d.Get("ensure_readiness").(bool) && strings.ToLower(resp.Data.Status) != "ready" {
		if ed, ok := ensureProjectIsReady(ctx, client, resp.Data.ID); !ok {
			return ed
		}
	}

	d.SetId(resp.Data.ID)
	return resourceDbtProjectRead(ctx, d, m)
}

func ensureProjectIsReady(ctx context.Context, client *fivetran.Client, projectId string) (diag.Diagnostics, bool) {
	var diags diag.Diagnostics
	for {
		s, errs, e := pollProjectStatus(ctx, client, projectId)
		if e != nil {
			deleteResp, err := client.NewDbtProjectDelete().DbtProjectID(projectId).Do(context.Background())
			if err != nil {
				return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("failed to cleanup after unsuccesful deletion; error: %v; code: %v; message: %v", err, deleteResp.Code, deleteResp.Message)), false
			}
			return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("unable to get status for dbt project: %v error: %v", projectId, err)), false
		}
		if s != "not_ready" {
			if s == "ready" {
				break
			} else {
				deleteResp, err := client.NewDbtProjectDelete().DbtProjectID(projectId).Do(context.Background())
				if err != nil {
					return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("failed to cleanup after unsuccesful deletion; error: %v; code: %v; message: %v", err, deleteResp.Code, deleteResp.Message)), false
				}
				return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("dbt project: %v has \"ERROR\" status after creation; errors: %v;", projectId, errs)), false
			}
		}

		if dl, ok := ctx.Deadline(); ok && time.Now().After(dl.Add(-time.Minute)) {
			// deadline will be exceeded on next iteration - it's time to cleanup
			deleteResp, err := client.NewDbtProjectDelete().DbtProjectID(projectId).Do(context.Background())
			if err != nil {
				return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("failed to cleanup after unsuccesful deletion; error: %v; code: %v; message: %v", err, deleteResp.Code, deleteResp.Message)), false
			}
			return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("project %v is stuck in \"NOT_READY\" status", projectId)), false
		}
		contextDelay(ctx, time.Second)
	}
	return diags, true
}

func pollProjectStatus(ctx context.Context, client *fivetran.Client, projectId string) (string, []string, error) {
	resp, err := client.NewDbtProjectDetails().DbtProjectID(projectId).Do(ctx)
	if err != nil {
		return "", []string{}, err
	}
	return strings.ToLower(resp.Data.Status), resp.Data.Errors, err
}

func resourceDbtProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := client.NewDbtProjectDetails().DbtProjectID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		// If the resource does not exist (404), inform Terraform. We want to immediately
		// return here to prevent further processing.
		if resp.Code == "404" {
			d.SetId("")
			return nil
		}
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	mapStringInterface := make(map[string]interface{})
	mapAddStr(mapStringInterface, "id", resp.Data.ID)
	mapAddStr(mapStringInterface, "group_id", resp.Data.GroupID)
	mapAddStr(mapStringInterface, "default_schema", resp.Data.DefaultSchema)
	mapAddStr(mapStringInterface, "dbt_version", resp.Data.DbtVersion)
	mapAddStr(mapStringInterface, "target_name", resp.Data.TargetName)
	mapAddStr(mapStringInterface, "type", resp.Data.Type)

	mapStringInterface["threads"] = resp.Data.Threads

	mapAddXString(mapStringInterface, "environment_vars", resp.Data.EnvironmentVars)

	upstreamConfig := make(map[string]interface{})
	mapAddStr(upstreamConfig, "git_remote_url", resp.Data.ProjectConfig.GitRemoteUrl)
	mapAddStr(upstreamConfig, "git_branch", resp.Data.ProjectConfig.GitBranch)
	mapAddStr(upstreamConfig, "folder_path", resp.Data.ProjectConfig.FolderPath)
	projectConfig := make([]interface{}, 0)
	mapStringInterface["project_config"] = append(projectConfig, upstreamConfig)

	mapAddStr(mapStringInterface, "created_at", resp.Data.CreatedAt)
	mapAddStr(mapStringInterface, "created_by_id", resp.Data.CreatedById)
	mapAddStr(mapStringInterface, "public_key", resp.Data.PublicKey)
	mapAddStr(mapStringInterface, "status", resp.Data.Status)

	if strings.ToLower(resp.Data.Status) == "ready" {
		modelsResp, err := getAllDbtModelsForProject(client, ctx, resp.Data.ID)
		if err != nil {
			return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, modelsResp.Code, modelsResp.Message))
		}
		mapAddXInterface(mapStringInterface, "models", flattenDbtModels(modelsResp))
	}

	for k, v := range mapStringInterface {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)
	return diags
}

func resourceDbtProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDbtProjectModify().DbtProjectID(d.Get("id").(string))

	if d.HasChange("dbt_version") {
		svc.DbtVersion(d.Get("dbt_version").(string))
	}
	if d.HasChange("target_name") {
		svc.TargetName(d.Get("target_name").(string))
	}
	if d.HasChange("threads") {
		svc.Threads(d.Get("threads").(int))
	}

	if d.HasChange("environment_vars") {
		vars := make([]string, 0)
		for _, varValue := range d.Get("environment_vars").(*schema.Set).List() {
			vars = append(vars, varValue.(string))
		}
		svc.EnvironmentVars(vars)
	}

	if d.HasChanges("project_config.0.git_branch", "project_config.0.folder_path") {
		projectConfig := fivetran.NewDbtProjectConfig()
		if d.HasChange("project_config.0.git_branch") {
			projectConfig.GitBranch(d.Get("project_config.0.git_branch").(string))
		}
		if d.HasChange("project_config.0.folder_path") {
			projectConfig.FolderPath(d.Get("project_config.0.folder_path").(string))
		}
		svc.ProjectConfig(projectConfig)
	}

	resp, err := svc.Do(ctx)

	// read and update state
	diags = resourceDbtProjectRead(ctx, d, m)

	if err != nil {
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}
	return diags
}

func resourceDbtProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := client.NewDbtProjectDelete().DbtProjectID(d.Get("id").(string)).Do(ctx)

	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}
	d.SetId("")
	return diags
}
