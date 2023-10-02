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
		Schema:      resourceTeamConnectorMembershipBase(true),
	}
}

func dataSourceTeamConnectorMembershipsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)
	teamId := d.Get("team_id").(string)

	connectors, err := dataSourceTeamConnectorMembershipsGet(client, ctx, teamId)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "team memberships service error", fmt.Sprintf("%v; code: %v", err, connectors.Code))
	}

	if err := d.Set("connector", dataSourceTeamConnectorMembershipsFlatten(&connectors)); err != nil {
		return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}
	
	d.SetId(teamId)

	return diags
}

// dataSourceTeamConnectorMembershipsFlatten receives a *fivetran.TeamConnectorMembershipsListResponse and returns a []interface{}
// containing the data type accepted by the "TeamConnectorMemberships" set.
func dataSourceTeamConnectorMembershipsFlatten(resp *fivetran.TeamConnectorMembershipsListResponse) []interface{} {
	if resp.Data.Items == nil {
		return make([]interface{}, 0)
	}

	memberships := make([]interface{}, len(resp.Data.Items))
	for i, v := range resp.Data.Items {
		membership := make(map[string]interface{})
		membership["connector_id"] = v.ConnectorId
		membership["role"] = v.Role
		membership["created_at"] = v.CreatedAt

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
