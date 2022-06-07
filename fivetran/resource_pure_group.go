package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePureGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePureGroupCreate,
		ReadContext:   resourcePureGroupRead,
		UpdateContext: resourcePureGroupUpdate,
		DeleteContext: resourcePureGroupDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"id":           {Type: schema.TypeString, Computed: true},
			"name":         {Type: schema.TypeString, Required: true},
			"created_at":   {Type: schema.TypeString, Computed: true},
			"creator":      {Type: schema.TypeString, Computed: true},
			"last_updated": {Type: schema.TypeString, Computed: true}, // internal
		},
	}
}

func resourcePureGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupCreate()

	resp, err := svc.Name(d.Get("name").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "create error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	groupID := resp.Data.ID

	groupCreator, err := resourceGroupUsersGetCreator(client, resp.Data.ID, ctx)
	if err != nil {
		// If resourceGroupGetCreator returns an error, the recently created group is deleted
		respDelete, errDelete := client.NewGroupDelete().GroupID(groupID).Do(ctx)
		if errDelete != nil {
			diags = newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, respDelete.Code, respDelete.Message))
		}

		return newDiagAppend(diags, diag.Error, "create error: groupCreator", fmt.Sprint(err))
	}
	if err := d.Set("creator", groupCreator); err != nil {
		// If d.Set returns an error, the recently created group is deleted
		respDelete, errDelete := client.NewGroupDelete().GroupID(groupID).Do(ctx)
		if errDelete != nil {
			diags = newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, respDelete.Code, respDelete.Message))
		}

		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	d.SetId(resp.Data.ID)
	resourcePureGroupRead(ctx, d, m)

	return diags
}

func resourcePureGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupDetails()

	groupID := d.Get("id").(string)
	svc.GroupID(groupID)

	resp, err := svc.Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "read error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.ID
	msi["name"] = resp.Data.Name
	msi["created_at"] = resp.Data.CreatedAt.String()
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	return diags
}

func resourcePureGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupModify()
	var change bool

	groupID := d.Get("id").(string)
	svc.GroupID(groupID)

	if d.HasChange("name") {
		svc.Name(d.Get("name").(string))
		change = true
	}

	if change {
		resp, err := svc.Do(ctx)
		if err != nil {
			// resourceGroupRead here makes sure the state is updated after a NewGroupModify error.
			diags = resourcePureGroupRead(ctx, d, m)
			return newDiagAppend(diags, diag.Error, "update error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
		}
	}

	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	return resourcePureGroupRead(ctx, d, m)
}

func resourcePureGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupDelete()

	resp, err := svc.GroupID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "delete error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	d.SetId("")

	return diags
}
