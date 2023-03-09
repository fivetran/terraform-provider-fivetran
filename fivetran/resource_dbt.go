package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDbt() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDbtCreate,
		ReadContext:   resourceDbtRead,
		UpdateContext: resourceDbtUpdate,
		DeleteContext: resourceDbtDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id":             {Type: schema.TypeString, Computed: true},
			"group_id":       {Type: schema.TypeString, Required: true, ForceNew: true},
			"created_at":     {Type: schema.TypeString, Computed: true},
			"public_key":     {Type: schema.TypeString, Computed: true},
			"git_remote_url": {Type: schema.TypeString, Computed: true},
			"git_branch":     {Type: schema.TypeString, Computed: true},
			"default_schema": {Type: schema.TypeString, Computed: true},
			"folder_path":    {Type: schema.TypeString, Computed: true},
			"target_name":    {Type: schema.TypeString, Computed: true},
			"last_updated":   {Type: schema.TypeString, Computed: true}, //internal
		},
	}
}

func resourceDbtCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDbtCreate()

	svc.GroupID(d.Get("group_id").(string))

	currentService := d.Get("service").(string)

	svc.Service(currentService)
	svc.DbtVersion(d.Get("dbt_version").(string))
	svc.GitRemoteUrl(d.Get("git_remote_url").(string))
	svc.GitBranch(d.Get("git_branch").(string))
	svc.DefaultSchema(d.Get("default_schema").(string))
	svc.FolderPath(d.Get("folder_path").(string))
	svc.TargetName(d.Get("target_name").(string))
	svc.Threads(strToInt(d.Get("threads")).(string))

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId(resp.Data.ID)
	resourceDbtRead(ctx, d, m)

	return diags
}

func resourceDbtRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostic
	client := m.(*fivetran.Client)

	resp, err := client.NewDbtDetails().DbtID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		// If the resource does not exist (404), inform Terraform. We want to immediately
		// return here to prevent further processing.
		if resp.Code == "404" {
			d.SetId("")
			return nil
		}
		return newDiagAppend(
			diags,
			diag.Error,
			"read error",
			fmt.Sprintf("%v; code %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stants for map string interface
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "group_id", resp.Data.GroupID)

	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())
	mapAddStr(msi, "created_by_id", resp.Data.CreatedById)
	mapAddStr(msi, "public_key", resp.Data.PublicKey)
	mapAddStr(msi, "git_remote_url", resp.Data.GitRemoteUrl)
	mapAddStr(msi, "git_branch", resp.Data.GitBranch)
	mapAddStr(msi, "default_schema", resp.Data.DefaultSchema)
	mapAddStr(msi, "folder_path", resp.Data.FolderPath)
	mapAddStr(msi, "target_name", resp.Data.TargetName)

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func resourceDbtRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostic
	client := m.(*fivetran.Client)

	resp, err := client.NewDbDetails().DbtID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		// If the resource does not exist (404), inform Terraform. We want to immediately
		// return here to prevent further processing

		if resp.Code == "404" {
			d.SetId("")
			return nil
		}
		return newDiagAppend(
			diags,
			diag.Error,
			"read error",
			fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stans for map string interface
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "group_id", resp.Data.GroupID)

	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())
	mapAddStr(msi, "created_by_id", resp.Data.CreatedById)
	mapAddStr(msi, "public_key", resp.Data.PublicKey)
	mapAddStr(msi, "git_remote_url", resp.Data.GitRemoteUrl)
	mapAddStr(msi, "git_branch", resp.Data.GitBranch)
	mapAddStr(msi, "default_schema", resp.Data.DefaultSchema)
	mapAddStr(msi, "folder_path", resp.Data.FolderPath)
	mapAddStr(msi, "target_name", resp.Data.TargetName)

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func resourceDbtUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewDbtModify()

	svc.DbtID(d.Get("id").(string))

	if d.HasChange("group_id") {
		svc.GroupID(d.Get("group_id").(string))
	}
	if d.HasChange("dbt_version") {
		svc.DbtVersion(d.Get("dbt_version").(string))
	}
	if d.HasChange("git_remote_url") {
		svc.GitRemoteUrl(d.Get("git_remote_url").(string))
	}
	if d.HasChange("git_branch") {
		svc.GitBranch(d.Get("git_branch").(string))
	}
	if d.HasChange("default_schema") {
		svc.DefaultSchema(d.Get("default_schema").(string))
	}
	if d.HasChange("folder_path") {
		svc.FolderPath(d.Get("folder_path").(string))
	}
	if d.HasChange("target_name") {
		svc.TargetName(d.Get("target_name").(string))
	}
	if d.HasChange("threads") {
		svc.Threads(strToInt(d.Get("threads").(string)))
	}

	resp, err := svc.Do(ctx)
	if err != nil {
		// resourceDbtRead here makes sure the state is updated after a NewDbtModify error.
		diags = resourceDbtRead(ctx, d, m)
		return newDiagAppend(
			diags,
			diag.Error,
			"update error",
			fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourceDbtRead(ctx, d, m)
}

func resourceDbtDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostic
	client := m.(*fivetran.Client)
	svc := client.NewConnectorDelete()

	resp, err := svc.DbtID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}
