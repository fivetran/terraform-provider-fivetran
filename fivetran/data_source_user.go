package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"given_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"family_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"verified": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"invited": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"picture": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"phone": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			// "role": &schema.Schema{ // commented until https://fivetran.height.app/T-109040 is fixed.
			// 	Type:     schema.TypeString,
			// 	Computed: true,
			// },
			"logged_in_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserDetails()

	id := d.Get("id").(string)

	resp, err := svc.UserID(id).Do(ctx)
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
	kvmap["email"] = resp.Data.Email
	kvmap["given_name"] = resp.Data.GivenName
	kvmap["family_name"] = resp.Data.FamilyName
	kvmap["verified"] = resp.Data.Verified
	kvmap["invited"] = resp.Data.Invited
	kvmap["picture"] = resp.Data.Picture
	kvmap["phone"] = resp.Data.Phone
	// kvmap["role"] = resp.Data.Role // commented until https://fivetran.height.app/T-109040 is fixed.
	kvmap["logged_in_at"] = resp.Data.LoggedInAt.String()
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
