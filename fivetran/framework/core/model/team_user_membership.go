package model

import (
    "context"

    "github.com/fivetran/go-fivetran"
    "github.com/fivetran/go-fivetran/teams"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type TeamUserMemberships struct {
    Id          types.String `tfsdk:"id"`
    TeamId      types.String `tfsdk:"team_id"`
    User        types.Set    `tfsdk:"user"`
}

func (d *TeamUserMemberships) ReadFromResponse(ctx context.Context, resp teams.TeamUserMembershipsListResponse) {
    elementType := map[string]attr.Type{
        "user_id":     types.StringType,
        "role":        types.StringType,
    }

    if resp.Data.Items == nil {
        d.User = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["user_id"] = types.StringValue(v.UserId)
        item["role"] = types.StringValue(v.Role)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }


    d.Id = d.TeamId
    d.User, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}

func (d *TeamUserMemberships) ReadFromSource(ctx context.Context, client *fivetran.Client, teamId string) (teams.TeamUserMembershipsListResponse, error) {
    var respNextCursor string
    var listResponse teams.TeamUserMembershipsListResponse
    limit := 1000

    svc := client.NewTeamUserMembershipsList()
    svc.TeamId(teamId)

    for {
        var err error
        var tmpResp teams.TeamUserMembershipsListResponse

        if respNextCursor == "" {
            tmpResp, err = svc.Limit(limit).Do(ctx)
        }

        if respNextCursor != "" {
            tmpResp, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
        }
        
        if err != nil {
            listResponse = teams.TeamUserMembershipsListResponse{}
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