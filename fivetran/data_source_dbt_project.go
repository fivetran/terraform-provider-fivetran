package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbtProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDbtProjectRead,
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
		},
	}
}

func dataSourceDbtProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostic {
	var diags diag.Diagnostic
	client := m.(*fivetran.Client)

	resp, err := client.NewDbtProjectDetails().DbtProjectID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service  error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for map string interface
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "group_id", resp.Data.GroupID)
	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())
	mapAddStr(msi, "public_key", resp.Data.PublicKey)
	mapAddStr(msi, "git_remote_url", resp.Data.GitRemoteUrl)
	mapAddStr(msi, "git_branch", resp.Data.GitBranch)
	mapAddStr(msi, "default_schema", resp.Data.DefaultSchema)
	mapAddStr(msi, "folder_path", resp.Data.FolderPath)
	mapAddStr(msi, "target_name", resp.Data.TargetName)
	mapAddStr(msi, "last_updated", resp.Data.lastUpdated)

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag, error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}
