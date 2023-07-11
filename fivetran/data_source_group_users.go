package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroupUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupUsersRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique identifier for the user within the account.",
			},
			"users": dataSourceGroupUsersSchemaUsers(),
		},
	}
}

func dataSourceGroupUsersSchemaUsers() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The unique identifier for the user within the account.",
				},
				"email": {
					Type:        schema.TypeString,
					Required:    true,
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
					Description: "The field indicates whether the user has verified their email address in the account creation process.",
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
					Required:    true,
					Description: "The group role that you would like to assign this new user to. Supported group roles: ‘Destination Administrator‘, ‘Destination Reviewer‘, ‘Destination Analyst‘, ‘Connector Creator‘, or a custom destination role",
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
	}
}

func dataSourceGroupUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	id := d.Get("id").(string)

	resp, err := dataSourceGroupUsersGetUsers(client, id, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("users", dataSourceGroupUsersFlattenUsers(&resp)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	d.SetId(id)

	msi := make(map[string]interface{})

	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	return diags
}

// dataSourceGroupUsersFlattenUsers receives a *fivetran.GroupListUsersResponse and returns a []interface{}
// containing the data type accepted by the "users" set.
func dataSourceGroupUsersFlattenUsers(resp *fivetran.GroupListUsersResponse) []interface{} {
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
		user["role"] = v.Role
		user["logged_in_at"] = v.LoggedInAt.String()
		user["created_at"] = v.CreatedAt.String()

		users[i] = user
	}

	return users
}

// dataSourceGroupUsersGetUsers gets the users list of a group. It handles limits and cursors.
func dataSourceGroupUsersGetUsers(client *fivetran.Client, id string, ctx context.Context) (fivetran.GroupListUsersResponse, error) {
	var resp fivetran.GroupListUsersResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.GroupListUsersResponse
		svc := client.NewGroupListUsers()
		if respNextCursor == "" {
			respInner, err = svc.GroupID(id).Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.GroupID(id).Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.GroupListUsersResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
