package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeamUserMemberships() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamUserMembershipsRead,
		Schema: map[string]*schema.Schema{
			"memberships": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: getTeamUserMembershipSchema(true),
				},
			},
		},
	}
}

func dataSourceTeamUserMembershipsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var result []interface{}

	client := m.(*fivetran.Client)

	teams, err := dataSourceTeamsGetTeams(client, ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", teams, teams.Code))
	}

	for _, v := range teams.Data.Items {
		cur_membership, err := dataSourceTeamUserMembershipsGet(client, ctx, v.Id)
		if err != nil {
			return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v", err, cur_membership.Code))
		}

		result = append(result, dataSourceTeamUserMembershipsFlatten(&cur_membership, v.Id)...)
	}

	if err := d.Set("memberships", result); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	// Enforces ID, there can't be two account-wide datasources
	d.SetId("0")

	return diags
}

// dataSourceTeamUserMembershipsFlatten receives a *fivetran.TeamUserMembershipsListResponse and returns a []interface{}
// containing the data type accepted by the "TeamUserMemberships" set.
func dataSourceTeamUserMembershipsFlatten(resp *fivetran.TeamUserMembershipsListResponse, teamId string) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	memberships := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		membership := make(map[string]interface{})
		membership["team_id"] = teamId
		membership["user_id"] = v.UserId
		membership["role"] = v.Role

		memberships[i] = membership
	}

	return memberships
}

// dataSourceTeamUserMembershipsGetTeamUserMemberships gets the memberships list of a User. It handles limits and cursors.
func dataSourceTeamUserMembershipsGet(client *fivetran.Client, ctx context.Context, teamId string) (fivetran.TeamUserMembershipsListResponse, error) {
	var resp fivetran.TeamUserMembershipsListResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.TeamUserMembershipsListResponse
		svc := client.NewTeamUserMembershipsList().TeamId(teamId)
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.TeamUserMembershipsListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
