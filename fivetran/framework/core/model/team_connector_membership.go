package model

import (
    "context"

    "github.com/fivetran/go-fivetran"
    "github.com/fivetran/go-fivetran/teams"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type TeamConnectionMemberships struct {
    Id          types.String `tfsdk:"id"`
    TeamId      types.String `tfsdk:"team_id"`
    Connection  types.Set    `tfsdk:"connector"`
}

func (d *TeamConnectionMemberships) ReadFromResponse(ctx context.Context, resp teams.TeamConnectionMembershipsListResponse) {
    elementType := map[string]attr.Type{
        "connector_id": types.StringType,
        "role":         types.StringType,
        "created_at":   types.StringType,
    }

    if resp.Data.Items == nil {
        d.Connection = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["connector_id"] = types.StringValue(v.ConnectionId)
        item["role"] = types.StringValue(v.Role)
        item["created_at"] = types.StringValue(v.CreatedAt)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Id = d.TeamId
    d.Connection, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}

func (d *TeamConnectionMemberships) ReadFromSource(ctx context.Context, client *fivetran.Client, teamId string) (teams.TeamConnectionMembershipsListResponse, error) {
    var respNextCursor string
    var listResponse teams.TeamConnectionMembershipsListResponse
    limit := 1000

    svc := client.NewTeamConnectionMembershipsList()
    svc.TeamId(teamId)

    for {
        var err error
        var tmpResp teams.TeamConnectionMembershipsListResponse

        if respNextCursor == "" {
            tmpResp, err = svc.Limit(limit).Do(ctx)
        }

        if respNextCursor != "" {
            tmpResp, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
        }
        
        if err != nil {
            listResponse = teams.TeamConnectionMembershipsListResponse{}
            return listResponse, err
        }

        listResponse.Data.Items = append(listResponse.Data.Items, tmpResp.Data.Items...)

        if tmpResp.Data.NextCursor == "" {
            break
        }

        respNextCursor = tmpResp.Data.NextCursor
    }

    return listResponse, nil
}