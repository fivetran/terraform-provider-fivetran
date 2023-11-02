package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRoles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRolesRead,
		Schema: map[string]*schema.Schema{
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set: func(v interface{}) int {
					return helpers.StringInt32Hash(v.(map[string]interface{})["name"].(string))
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role name",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The role description",
						},
						"is_custom": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "TypeBool",
						},
						"scope": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "Defines the list of resources the role manages. Supported values: ACCOUNT, DESTINATION, CONNECTOR, and TEAM",
						},
					},
				},
			},
		},
	}
}

func dataSourceRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := dataSourceRolesGet(client, ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", err, resp.Code))
	}

	if err := d.Set("roles", dataSourceRolesFlat(&resp)); err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID, there can't be two account-wide datasources
	d.SetId("0")

	return diags
}

// dataSourceRolesFlat receives a *fivetran.RolesListResponse and returns a []interface{}
// containing the data type accepted by the "roles" set.
func dataSourceRolesFlat(resp *fivetran.RolesListResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	roles := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		role := make(map[string]interface{})
		role["name"] = v.Name
		role["description"] = v.Description
		role["is_custom"] = v.IsCustom
		role["scope"] = v.Scope

		roles[i] = role
	}

	return roles
}

// dataSourceRolesGet gets the list of a roles. It handles limits and cursors.
func dataSourceRolesGet(client *fivetran.Client, ctx context.Context) (fivetran.RolesListResponse, error) {
	var resp fivetran.RolesListResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.RolesListResponse
		svc := client.NewRolesList()
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.RolesListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
