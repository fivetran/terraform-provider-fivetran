package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"users": {
				Type: schema.TypeSet,
				// Uncomment Optional:true, before re-generating docs
				//Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier for the user within the Fivetran system.",
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
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
				},
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := dataSourceUsersGetUsers(client, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("users", dataSourceUsersFlattenUsers(&resp)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID, there can't be two account-wide datasources
	d.SetId("0")

	return diags
}

// dataSourceUsersFlattenUsers receives a *fivetran.UsersListResponse and returns a []interface{}
// containing the data type accepted by the "users" set.
func dataSourceUsersFlattenUsers(resp *fivetran.UsersListResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	users := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		user := make(map[string]interface{})
		user["id"] = v.ID
		user["email"] = v.Email
		user["given_name"] = v.GivenName
		user["family_name"] = v.FamilyName
		user["verified"] = v.Verified
		user["invited"] = v.Invited
		user["picture"] = v.Picture
		user["phone"] = v.Phone
		user["logged_in_at"] = v.LoggedInAt.String()
		user["created_at"] = v.CreatedAt.String()

		users[i] = user
	}

	return users
}

// dataSourceGroupUsersGetUsers gets the users list of a group. It handles limits and cursors.
func dataSourceUsersGetUsers(client *fivetran.Client, ctx context.Context) (fivetran.UsersListResponse, error) {
	var resp fivetran.UsersListResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.UsersListResponse
		svc := client.NewUsersList()
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.UsersListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
