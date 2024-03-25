package model

import (
    "context"
    "time"

    "github.com/fivetran/go-fivetran"
    "github.com/fivetran/go-fivetran/groups"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type GroupUser struct {
	Id   		   types.String `tfsdk:"id"`
	GroupId        types.String `tfsdk:"group_id"`
	LastUpdated    types.String `tfsdk:"last_updated"`
    User   		   types.Set    `tfsdk:"user"`
}

func (d *GroupUser) ReadFromResponse(ctx context.Context, resp groups.GroupListUsersResponse) {
    elementType := map[string]attr.Type{
        "id":           types.StringType,
        "email":        types.StringType,
		"role":         types.StringType,
    }

    if resp.Data.Items == nil {
        d.User = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    users := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        user := map[string]attr.Value{}
		user["id"] = types.StringValue(v.ID)
		user["email"] = types.StringValue(v.Email)
		user["role"] = types.StringValue(v.Role)

        objectValue, _ := types.ObjectValue(elementType, user)
        users = append(users, objectValue)
    }

    d.Id = d.GroupId
    d.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
    d.User, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, users)
}

func (d *GroupUser) ReadFromSource(ctx context.Context, client *fivetran.Client, groupId string) (groups.GroupListUsersResponse, error) {
    var respNextCursor string
    var listResponse groups.GroupListUsersResponse
    limit := 1000

    svc := client.NewGroupListUsers()

    for {
        var err error
        var tmpResp groups.GroupListUsersResponse

        if respNextCursor == "" {
            tmpResp, err = svc.Limit(limit).GroupID(groupId).Do(ctx)
        }

        if respNextCursor != "" {
            tmpResp, err = svc.Limit(limit).GroupID(groupId).Cursor(respNextCursor).Do(ctx)
        }
        
        if err != nil {
            listResponse = groups.GroupListUsersResponse{}
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