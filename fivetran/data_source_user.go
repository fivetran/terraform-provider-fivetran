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
	c := m.(*fivetran.Client)
	s := c.NewUserDetails()

	id := d.Get("id").(string)
	user, err := s.UserID(id).Do(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "service error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, user.Code, user.Message),
		})
		return diags
	}

	kvmap := make(map[string]interface{})
	kvmap["id"] = user.Data.ID
	kvmap["email"] = user.Data.Email
	kvmap["given_name"] = user.Data.GivenName
	kvmap["family_name"] = user.Data.FamilyName
	kvmap["verified"] = user.Data.Verified
	kvmap["invited"] = user.Data.Invited
	kvmap["picture"] = user.Data.Picture
	kvmap["phone"] = user.Data.Phone
	// kvmap["role"] = user.Data.Role // commented until https://fivetran.height.app/T-109040 is fixed.
	kvmap["logged_in_at"] = user.Data.LoggedInAt.String()
	kvmap["created_at"] = user.Data.CreatedAt.String()

	if err := set(d, kvmap); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "set error",
			Detail:   fmt.Sprint(err),
		})
	}

	d.SetId(user.Data.ID)

	return diags
}
