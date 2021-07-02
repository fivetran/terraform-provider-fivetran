package fivetran

// WIP, not reviewed after 02/07/2021 yet.

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupRead,
		Schema: map[string]*schema.Schema{
			"id":         {Type: schema.TypeString, Required: true},
			"name":       {Type: schema.TypeString, Computed: true},
			"created_at": {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupDetails()

	id := d.Get("id").(string)

	resp, err := svc.GroupID(id).Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "service error",
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

	d.SetId(resp.Data.ID)

	return diags
}
