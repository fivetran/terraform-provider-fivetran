package model

import (
    "context"

    "github.com/fivetran/go-fivetran"
    "github.com/fivetran/go-fivetran/teams"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type Team struct {
    Id              types.String `tfsdk:"id"`
    Name            types.String `tfsdk:"name"`
    Description     types.String `tfsdk:"description"`
    Role            types.String `tfsdk:"role"`
}

func (d *Team) ReadFromResponse(ctx context.Context, resp teams.TeamsDetailsResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.Name = types.StringValue(resp.Data.Name)
    d.Description = types.StringValue(resp.Data.Description)
    d.Role = types.StringValue(resp.Data.Role)
}

func (d *Team) ReadFromCreateResponse(ctx context.Context, resp teams.TeamsCreateResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.Name = types.StringValue(resp.Data.Name)
    d.Description = types.StringValue(resp.Data.Description)
    d.Role = types.StringValue(resp.Data.Role)
}

func (d *Team) ReadFromModifyResponse(ctx context.Context, resp teams.TeamsModifyResponse) {
    d.Id = types.StringValue(resp.Data.Id)
    d.Name = types.StringValue(resp.Data.Name)
    d.Description = types.StringValue(resp.Data.Description)
    d.Role = types.StringValue(resp.Data.Role)
}

/*          */
type TeamConnectorMemberships struct {
    TeamId      types.String `tfsdk:"team_id"`
    Connector   types.Set    `tfsdk:"connector"`
}

func (d *TeamConnectorMemberships) ReadFromResponse(ctx context.Context, resp teams.TeamConnectorMembershipsListResponse) {
    elementType := map[string]attr.Type{
        "connector_id": types.StringType,
        "role":         types.StringType,
        "created_at":   types.StringType,
    }

    if resp.Data.Items == nil {
        d.Connector = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["connector_id"] = types.StringValue(v.ConnectorId)
        item["role"] = types.StringValue(v.Role)
        item["created_at"] = types.StringValue(v.CreatedAt)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }


    d.Connector, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}

func (d *TeamConnectorMemberships) ReadFromSource(ctx context.Context, client *fivetran.Client, teamId string) (teams.TeamConnectorMembershipsListResponse, error) {
    var respNextCursor string
    var listResponse teams.TeamConnectorMembershipsListResponse
    limit := 1000

    svc := client.NewTeamConnectorMembershipsList()
    svc.TeamId(teamId)

    for {
        var err error
        var tmpResp teams.TeamConnectorMembershipsListResponse

        if respNextCursor == "" {
            tmpResp, err = svc.Limit(limit).Do(ctx)
        }

        if respNextCursor != "" {
            tmpResp, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
        }
        
        if err != nil {
            listResponse = teams.TeamConnectorMembershipsListResponse{}
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

/*          */
type TeamGroupMemberships struct {
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

/*          */
type TeamUserMemberships struct {
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
