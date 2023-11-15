package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier for the user within the Fivetran system.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address that the user has associated with their user profile.",
			},
			"given_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The first name of the user.",
			},
			"family_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last name of the user.",
			},
			"verified": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The field indicates whether the user has verified their email address in the account creation process.",
			},
			"invited": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The field indicates whether the user has been invited to your account.",
			},
			"picture": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user's avatar as a URL link (for example, 'http://mycompany.com/avatars/john_white.png') or base64 data URI (for example, 'data:image/png;base64,aHR0cDovL215Y29tcGFueS5jb20vYXZhdGFycy9qb2huX3doaXRlLnBuZw==')",
			},
			"phone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The phone number of the user.",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The role that you would like to assign to the user",
			},
			"logged_in_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last time that the user has logged into their Fivetran account.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp that the user created their Fivetran account",
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewUserDetails()

	resp, err := svc.UserID(d.Get("id").(string)).Do(ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
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
	msi["role"] = resp.Data.Role
	msi["logged_in_at"] = resp.Data.LoggedInAt.String()
	msi["created_at"] = resp.Data.CreatedAt.String()
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}
