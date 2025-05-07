package model

import (
    "context"

    "github.com/fivetran/go-fivetran"
    "github.com/fivetran/go-fivetran/users"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type UserConnectionMemberships struct {
    Id            types.String `tfsdk:"id"`
    Connections   types.Set    `tfsdk:"Connections"`
}

func (d *UserConnectionMemberships) ReadFromResponse(ctx context.Context, resp users.UserConnectionMembershipsListResponse) {
    elementType := map[string]attr.Type{
        "connection_id": types.StringType,
        "role":          types.StringType,
        "created_at":    types.StringType,
    }

    if resp.Data.Items == nil {
        d.Connections = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["connection_id"] = types.StringValue(v.ConnectionId)
        item["role"] = types.StringValue(v.Role)
        item["created_at"] = types.StringValue(v.CreatedAt)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Connections, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}

func (d *UserConnectionMemberships) ReadFromSource(ctx context.Context, client *fivetran.Client, userId string) (users.UserConnectionMembershipsListResponse, error) {
    var respNextCursor string
    var listResponse users.UserConnectionMembershipsListResponse
    limit := 1000

    svc := client.NewUserConnectionMembershipsList()
    svc.UserId(userId)

    for {
        var err error
        var tmpResp users.UserConnectionMembershipsListResponse

        if respNextCursor == "" {
            tmpResp, err = svc.Limit(limit).Do(ctx)
        }

        if respNextCursor != "" {
            tmpResp, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
        }
        
        if err != nil {
            listResponse = users.UserConnectionMembershipsListResponse{}
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