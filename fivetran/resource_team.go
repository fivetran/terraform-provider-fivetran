package fivetran

import (
	"context"
	"fmt"

	fivetran "github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: 	   getTeamSchema(false),
	}
}

func getTeamSchema(datasource bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Required:    datasource,
			Computed:    !datasource,
			Description: "The unique identifier for the team within your account.",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    !datasource,
			Computed:    datasource,
			Description: "The name of the team within your account.",
		},
		"description": {
			Type:        schema.TypeString,
			Required:    !datasource,
			Computed:    datasource,
			Description: "The description of the team within your account.",
		},
		"role": {
			Type:        schema.TypeString,
			Required:    !datasource,
			Computed:    datasource,
			Description: "The account role of the team.",
		},
	}
}

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	svcAcc := client.NewTeamsCreate()
	svcAcc.Name(d.Get("name").(string))
	svcAcc.Role(d.Get("role").(string))
	svcAcc.Description(d.Get("description").(string))

	resp, err := svcAcc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v", err, resp.Code))
	}

	d.SetId(resp.Data.Id)

	resourceTeamRead(ctx, d, m)

	return diags
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewTeamsDetails()

	svc.TeamId(d.Get("id").(string)).Do(ctx)

	resp, err := svc.Do(ctx)
	if err != nil {
		// If the resource does not exist (404), inform Terraform. We want to immediately
		// return here to prevent further processing.
		if resp.Code == "404" {
			d.SetId("")
			return nil
		}
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v", err, resp.Code))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.Id
	msi["name"] = resp.Data.Name
	msi["description"] = resp.Data.Description
	msi["role"] = resp.Data.Role

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.Id)

	return diags
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	svc := client.NewTeamsModify()

	svc.TeamId(d.Get("id").(string))

	if d.HasChange("name") {
		svc.Name(d.Get("name").(string))
	}

	if d.HasChange("description") {
		svc.Description(d.Get("description").(string))
	}

	if d.HasChange("role") {
		svc.Role(d.Get("role").(string))
	}

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v", err, resp.Code))
	}		
	
	return resourceTeamRead(ctx, d, m)
}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewTeamsDelete()

	resp, err := svc.TeamId(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}
