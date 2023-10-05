package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/dbt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbtProjects() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDbtProjectsRead,
		Schema: map[string]*schema.Schema{
			"projects": dataSourceDbtProjectsSchema(),
		},
	}
}

func dataSourceDbtProjectsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
		Set: func(v interface{}) int {
			return stringInt32Hash(v.(map[string]interface{})["id"].(string))
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The unique identifier for the dbt project within the Fivetran system.",
				},
				"group_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The name of the group within your account related to the project.",
				},
				"created_at": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The timestamp of when the project was created in your account.",
				},
				"created_by_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The unique identifier for the User within the Fivetran system who created the DBT Project.",
				},
			},
		},
	}
}

func dataSourceDbtProjectsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := dataSourceDbtProjectsGetAllProjects(client, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	if err := d.Set("projects", dataSourceGroupsFlattenDbtProjects(resp)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID
	d.SetId("0")

	return diags
}

// dataSourceGroupsGetGroups gets the groups list. It handles limits and cursors.
func dataSourceDbtProjectsGetAllProjects(client *fivetran.Client, ctx context.Context) (dbt.DbtProjectsListResponse, error) {
	var resp dbt.DbtProjectsListResponse
	var respNextCursor string

	for {
		var err error
		var respInner dbt.DbtProjectsListResponse
		svc := client.NewDbtProjectsList()
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return dbt.DbtProjectsListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}

func dataSourceGroupsFlattenDbtProjects(response dbt.DbtProjectsListResponse) []interface{} {
	result := make([]interface{}, 0)

	for _, prj := range response.Data.Items {
		prjMap := make(map[string]interface{})
		prjMap["id"] = prj.ID
		prjMap["group_id"] = prj.GroupId
		prjMap["created_at"] = prj.CreatedAt
		prjMap["created_by_id"] = prj.CreatedById
		result = append(result, prjMap)
	}

	return result
}
