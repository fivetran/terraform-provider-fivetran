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
		Schema:      resourceTeamUserMembershipBase(true),
	}
}

func dataSourceTeamUserMembershipsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	teamId := d.Get("team_id").(string)

	users, err := dataSourceTeamUserMembershipsGet(client, ctx, teamId)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "team memberships service error", fmt.Sprintf("%v; code: %v", err, users.Code))
	}

	if err := d.Set("user", dataSourceTeamUserMembershipsFlatten(&users)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}
	
	d.SetId(teamId)

	return diags
}

// dataSourceTeamUserMembershipsFlatten receives a *fivetran.TeamUserMembershipsListResponse and returns a []interface{}
// containing the data type accepted by the "TeamUserMemberships" set.
func dataSourceTeamUserMembershipsFlatten(resp *fivetran.TeamUserMembershipsListResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	memberships := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		membership := make(map[string]interface{})
		membership["user_id"] = v.UserId
		membership["role"] = v.Role

		memberships[i] = membership
	}

	return memberships
}

// dataSourceTeamUserMembershipsGetTeamUserMemberships gets the memberships list of a user. It handles limits and cursors.
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
