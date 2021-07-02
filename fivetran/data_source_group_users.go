package fivetran

// WIP, not reviewed after 02/07/2021 yet.

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
			"id": {Type: schema.TypeString, Required: true},
			"users": {Type: schema.TypeSet, Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":           {Type: schema.TypeString, Computed: true},
						"email":        {Type: schema.TypeString, Computed: true},
						"given_name":   {Type: schema.TypeString, Computed: true},
						"family_name":  {Type: schema.TypeString, Computed: true},
						"verified":     {Type: schema.TypeBool, Computed: true},
						"invited":      {Type: schema.TypeBool, Computed: true},
						"picture":      {Type: schema.TypeString, Computed: true},
						"phone":        {Type: schema.TypeString, Computed: true},
						"logged_in_at": {Type: schema.TypeString, Computed: true},
						"created_at":   {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceGroupUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	svc := client.NewGroupListUsers()

	id := d.Get("id").(string)

	resp, err := dataSourceGroupUsersGetUsers(svc, id, ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "service error",
			Detail:   fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message),
		})
		return diags
	}

	d.Set("users", flattenGroupUsers(&resp))

	d.SetId(id)

	return diags
}

func flattenGroupUsers(resp *fivetran.GroupListUsersResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	users := make([]interface{}, len(resp.Data.Items), len(resp.Data.Items))

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

// dataSourceGroupUsersGetUsers gets the list of users of a group. It handles limits and cursors.
func dataSourceGroupUsersGetUsers(svc *fivetran.GroupListUsersService, id string, ctx context.Context) (fivetran.GroupListUsersResponse, error) {
	var resp fivetran.GroupListUsersResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.GroupListUsersResponse
		if respNextCursor == "" {
			respInner, err = svc.GroupID(id).Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.GroupID(id).Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.GroupListUsersResponse{}, err
		}

		for _, item := range respInner.Data.Items {
			resp.Data.Items = append(resp.Data.Items, item)
		}

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
