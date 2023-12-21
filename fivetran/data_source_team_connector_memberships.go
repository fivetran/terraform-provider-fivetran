package fivetran

import (
	"context"
	"fmt"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/teams"
	"github.com/fivetran/terraform-provider-fivetran/modules/helpers"
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
		return helpers.NewDiagAppend(diags, diag.Error, "team memberships service error", fmt.Sprintf("%v; code: %v", err, connectors.Code))
	}

	if err := d.Set("connector", dataSourceTeamConnectorMembershipsFlatten(&connectors)); err != nil {
		return helpers.NewDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
	}

	d.SetId(teamId)

	return diags
}

// dataSourceTeamConnectorMembershipsFlatten receives a *teams.TeamConnectorMembershipsListResponse and returns a []interface{}
// containing the data type accepted by the "TeamConnectorMemberships" set.
func dataSourceTeamConnectorMembershipsFlatten(resp *teams.TeamConnectorMembershipsListResponse) []interface{} {
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
func dataSourceTeamConnectorMembershipsGet(client *fivetran.Client, ctx context.Context, teamId string) (teams.TeamConnectorMembershipsListResponse, error) {
	var resp teams.TeamConnectorMembershipsListResponse
	var respNextCursor string

	for {
		var err error
		var respInner teams.TeamConnectorMembershipsListResponse
		svc := client.NewTeamConnectorMembershipsList().TeamId(teamId)
		if respNextCursor == "" {
			respInner, err = svc.Limit(limit).Do(ctx)
		}
		if respNextCursor != "" {
			respInner, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
		}
		if err != nil {
			return teams.TeamConnectorMembershipsListResponse{}, err
		}

		resp.Data.Items = append(resp.Data.Items, respInner.Data.Items...)

		if respInner.Data.NextCursor == "" {
			break
		}

		respNextCursor = respInner.Data.NextCursor
	}

	return resp, nil
}
