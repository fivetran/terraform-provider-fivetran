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
			"id":          {Type: schema.TypeString, Required: true},
			"email":       {Type: schema.TypeString, Computed: true},
			"given_name":  {Type: schema.TypeString, Computed: true},
			"family_name": {Type: schema.TypeString, Computed: true},
			"verified":    {Type: schema.TypeBool, Computed: true},
			"invited":     {Type: schema.TypeBool, Computed: true},
			"picture":     {Type: schema.TypeString, Computed: true},
			"phone":       {Type: schema.TypeString, Computed: true},
			// "role":         {Type: schema.TypeString, Computed: true}, // commented until T-109040 is fixed.
			"logged_in_at": {Type: schema.TypeString, Computed: true},
			"created_at":   {Type: schema.TypeString, Computed: true},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserDetails()

	resp, err := svc.UserID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	msi["id"] = resp.Data.ID
	msi["email"] = resp.Data.Email
	msi["given_name"] = resp.Data.GivenName
	msi["family_name"] = resp.Data.FamilyName
	msi["verified"] = resp.Data.Verified
	msi["invited"] = resp.Data.Invited
	msi["picture"] = resp.Data.Picture
	msi["phone"] = resp.Data.Phone
	// msi["role"] = resp.Data.Role // T-109040 is fixed.
	msi["logged_in_at"] = resp.Data.LoggedInAt.String()
	msi["created_at"] = resp.Data.CreatedAt.String()
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}
