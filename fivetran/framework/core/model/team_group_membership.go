package model

import (
    "context"

    "github.com/fivetran/go-fivetran"
    "github.com/fivetran/go-fivetran/teams"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type TeamGroupMemberships struct {
    Id          types.String `tfsdk:"id"`
    TeamId      types.String `tfsdk:"team_id"`
    Group       types.Set    `tfsdk:"group"`
}

func (d *TeamGroupMemberships) ReadFromResponse(ctx context.Context, resp teams.TeamGroupMembershipsListResponse) {
    elementType := map[string]attr.Type{
        "group_id":     types.StringType,
        "role":         types.StringType,
        "created_at":   types.StringType,
    }

    if resp.Data.Items == nil {
        d.Group = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["group_id"] = types.StringValue(v.GroupId)
        item["role"] = types.StringValue(v.Role)
        item["created_at"] = types.StringValue(v.CreatedAt)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Id = d.TeamId
    d.Group, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}

func (d *TeamGroupMemberships) ReadFromSource(ctx context.Context, client *fivetran.Client, teamId string) (teams.TeamGroupMembershipsListResponse, error) {
    var respNextCursor string
    var listResponse teams.TeamGroupMembershipsListResponse
    limit := 1000

    svc := client.NewTeamGroupMembershipsList()
    svc.TeamId(teamId)

    for {
        var err error
        var tmpResp teams.TeamGroupMembershipsListResponse

        if respNextCursor == "" {
            tmpResp, err = svc.Limit(limit).Do(ctx)
        }

        if respNextCursor != "" {
            tmpResp, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
        }
        
        if err != nil {
            listResponse = teams.TeamGroupMembershipsListResponse{}
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