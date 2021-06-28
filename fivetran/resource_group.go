package fivetran

import (
	"context"
	"fmt"
	"time"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"created_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			// "users": &schema.Schema{ // new code, CONTINUE HERE
			// 	Type:     schema.TypeSet,
			// 	Optional: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"user_id": {
			// 				Type:     schema.TypeString,
			// 				Optional: true,
			// 			},
			// 		},
			// 	},
			// },
			"last_updated": &schema.Schema{ // internal
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupCreate()

	name := d.Get("name").(string)
	svc.Name(name)

	resp, err := svc.Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "create error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message),
		})
		return diags
	}

	d.SetId(resp.Data.ID)

	resourceUserRead(ctx, d, m)

	return diags
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupDetails()

	id := d.Get("id").(string)
	svc.GroupID(id)

	resp, err := svc.Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "read error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message),
		})
		return diags
	}

	kvmap := make(map[string]interface{})
	kvmap["id"] = resp.Data.ID
	kvmap["name"] = resp.Data.Name
	kvmap["created_at"] = resp.Data.CreatedAt.String()

	if err := set(d, kvmap); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "set error",
			Detail:   fmt.Sprint(err),
		})
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var change bool
	client := m.(*fivetran.Client)
	svc := client.NewGroupModify()

	id := d.Get("id").(string)
	svc.GroupID(id)

	if d.HasChange("name") {
		name := d.Get("name").(string)
		svc.Name(name)
		change = true
	}

	if change {
		resp, err := svc.Do(ctx)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "update error",
				Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message),
			})
			return diags
		}

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceGroupRead(ctx, d, m)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupDelete()

	id := d.Get("id").(string)
	svc.GroupID(id)

	resp, err := svc.Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "delete error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message),
		})
		return diags
	}

	d.SetId("")

	return diags
}
