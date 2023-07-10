package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGroupsRead,
		Schema: map[string]*schema.Schema{
			"groups": dataSourceGroupSchemaGroups(),
		},
	}
}

func dataSourceGroupSchemaGroups() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeSet,
		// Uncomment `Optional: true,` before re-generating docs
		// Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The unique identifier for the group within the Fivetran system.",
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The name of the group within your account.",
				},
				"created_at": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The timestamp of when the group was created in your account.",
				},
			},
		},
	}
}

func dataSourceGroupsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := dataSourceGroupsGetGroups(client, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("groups", dataSourceGroupsFlattenGroups(&resp)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID
	d.SetId("0")

	return diags
}

// dataSourceGroupsFlattenGroups receives a *fivetran.GroupsListResponse and returns a []interface{}
// containing the data type accepted by the "groups" set.
func dataSourceGroupsFlattenGroups(resp *fivetran.GroupsListResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	groups := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		group := make(map[string]interface{})
		group["id"] = v.ID
		group["name"] = v.Name
		group["created_at"] = v.CreatedAt.String()

		groups[i] = group
	}

	return groups
}

// dataSourceGroupsGetGroups gets the groups list. It handles limits and cursors.
func dataSourceGroupsGetGroups(client *fivetran.Client, ctx context.Context) (fivetran.GroupsListResponse, error) {
	var resp fivetran.GroupsListResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.GroupsListResponse
		svc := client.NewGroupsList()
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.GroupsListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
