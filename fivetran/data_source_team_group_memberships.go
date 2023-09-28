package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeamGroupMemberships() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamGroupMembershipsRead,
		Schema:      resourceTeamGroupMembershipBase(true),
	}
}

func dataSourceTeamGroupMembershipsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	teamId := d.Get("team_id").(string)

	groups, err := dataSourceTeamGroupMembershipsGet(client, ctx, teamId)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "team memberships service error", fmt.Sprintf("%v; code: %v", err, groups.Code))
	}

	if err := d.Set("group", dataSourceTeamGroupMembershipsFlatten(&groups)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}
	
	d.SetId(teamId)

	return diags
}

// dataSourceTeamGroupMembershipsFlatten receives a *fivetran.TeamGroupMembershipsListResponse and returns a []interface{}
// containing the data type accepted by the "TeamGroupMemberships" set.
func dataSourceTeamGroupMembershipsFlatten(resp *fivetran.TeamGroupMembershipsListResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	memberships := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		membership := make(map[string]interface{})
		membership["group_id"] = v.GroupId
		membership["role"] = v.Role
		membership["created_at"] = v.CreatedAt

		memberships[i] = membership
	}

	return memberships
}

// dataSourceTeamGroupMembershipsGetTeamGroupMemberships gets the memberships list of a group. It handles limits and cursors.
func dataSourceTeamGroupMembershipsGet(client *fivetran.Client, ctx context.Context, teamId string) (fivetran.TeamGroupMembershipsListResponse, error) {
	var resp fivetran.TeamGroupMembershipsListResponse
	var respNextCursor string

	for {
		var err error
		var respInner fivetran.TeamGroupMembershipsListResponse
		svc := client.NewTeamGroupMembershipsList().TeamId(teamId)
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return fivetran.TeamGroupMembershipsListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
