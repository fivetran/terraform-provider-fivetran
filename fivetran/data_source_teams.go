package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeams() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamsRead,
		Schema: map[string]*schema.Schema{
			"teams": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Set: func(v interface{}) int {
					return helpers.StringInt32Hash(v.(map[string]interface{})["id"].(string))
				},
				Elem: &schema.Resource{
					Schema: getTeamSchema(true),
				},
			},
		},
	}
}

func dataSourceTeamsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := dataSourceTeamsGetTeams(client, ctx)
	if err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", err, resp.Code))
	}

	if err := d.Set("teams", dataSourceTeamsFlattenTeams(&resp)); err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID, there can't be two account-wide datasources
	d.SetId("0")

	return diags
}

// dataSourceTeamsFlattenTeams receives a *fivetran.TeamsListResponse and returns a []interface{}
// containing the data type accepted by the "teams" set.
func dataSourceTeamsFlattenTeams(resp *fivetran.TeamsListResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	teams := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		team := make(map[string]interface{})
		team["id"] = v.Id
		team["name"] = v.Name
		team["description"] = v.Description
		team["role"] = v.Role

		teams[i] = team
	}

	return teams
}

// dataSourceTeamsGetTeams gets the teams list of a group. It handles limits and cursors.
func dataSourceTeamsGetTeams(client *fivetran.Client, ctx context.Context) (fivetran.TeamsListResponse, error) {
	var resp fivetran.TeamsListResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.TeamsListResponse
		svc := client.NewTeamsList()
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.TeamsListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
