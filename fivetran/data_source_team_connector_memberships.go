package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeamConnectorMemberships() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamConnectorMembershipsRead,
		Schema: map[string]*schema.Schema{
			"memberships": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getTeamConnectorMembershipSchema(true),
				},
			},
		},
	}
}

func dataSourceTeamConnectorMembershipsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var result []interface{}

	client := m.(*fivetran.Client)

	teams, err := dataSourceTeamsGetTeams(client, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", teams, teams.Code))
	}

	for _, v := range teams.Data.Items {
		cur_membership, err := dataSourceTeamConnectorMembershipsGet(client, ctx, v.Id)
		if err != nil {
			return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", err, cur_membership.Code))
		}

		result = append(result, dataSourceTeamConnectorMembershipsFlatten(&cur_membership, v.Id)...)
	}

	if err := d.Set("memberships", result); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID, there can't be two account-wide datasources
	d.SetId("0")

	return diags
}

// dataSourceTeamConnectorMembershipsFlatten receives a *fivetran.TeamConnectorMembershipsListResponse and returns a []interface{}
// containing the data type accepted by the "TeamConnectorMemberships" set.
func dataSourceTeamConnectorMembershipsFlatten(resp *fivetran.TeamConnectorMembershipsListResponse, teamId string) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	memberships := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		membership := make(map[string]interface{})
		membership["team_id"] = teamId
		membership["connector_id"] = v.ConnectorId
		membership["role"] = v.Role

		memberships[i] = membership
	}

	return memberships
}

// dataSourceTeamConnectorMembershipsGetTeamConnectorMemberships gets the memberships list of a group. It handles limits and cursors.
func dataSourceTeamConnectorMembershipsGet(client *fivetran.Client, ctx context.Context, teamId string) (fivetran.TeamConnectorMembershipsListResponse, error) {
	var resp fivetran.TeamConnectorMembershipsListResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.TeamConnectorMembershipsListResponse
		svc := client.NewTeamConnectorMembershipsList().TeamId(teamId)
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.TeamConnectorMembershipsListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
