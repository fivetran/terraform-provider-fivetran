package model

import (
    "context"

    "github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/users"
	"github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type User struct {
	ID         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	GivenName  types.String `tfsdk:"given_name"`
	FamilyName types.String `tfsdk:"family_name"`
	Verified   types.Bool   `tfsdk:"verified"`
	Invited    types.Bool   `tfsdk:"invited"`
	Picture    types.String `tfsdk:"picture"`
	Phone      types.String `tfsdk:"phone"`
	Role       types.String `tfsdk:"role"`
	LoggedInAt types.String `tfsdk:"logged_in_at"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func (d *User) ReadFromResponse(resp users.UserDetailsResponse) {
	d.ID = types.StringValue(resp.Data.ID)
	d.Email = types.StringValue(resp.Data.Email)
	d.FamilyName = types.StringValue(resp.Data.FamilyName)
	d.GivenName = types.StringValue(resp.Data.GivenName)

	d.Role = types.StringValue(resp.Data.Role)
	d.Verified = types.BoolValue(*resp.Data.Verified)
	d.Invited = types.BoolValue(*resp.Data.Invited)
	d.LoggedInAt = types.StringValue(resp.Data.LoggedInAt.String())
	d.CreatedAt = types.StringValue(resp.Data.CreatedAt.String())

	if resp.Data.Phone == "" {
		d.Phone = types.StringNull()
	} else {
		d.Phone = types.StringValue(resp.Data.Phone)
	}

	if resp.Data.Picture == "" {
		d.Picture = types.StringNull()
	} else {
		d.Picture = types.StringValue(resp.Data.Picture)
	}
}

/*          */
type UserConnectorMemberships struct {
    UserId      types.String `tfsdk:"user_id"`
    Connector   types.Set    `tfsdk:"connector"`
}

func (d *UserConnectorMemberships) ReadFromResponse(ctx context.Context, resp users.UserConnectionMembershipsListResponse) {
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
        item["connector_id"] = types.StringValue(v.ConnectionId)
        item["role"] = types.StringValue(v.Role)
        item["created_at"] = types.StringValue(v.CreatedAt)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }


    d.Connector, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}

func (d *UserConnectorMemberships) ReadFromSource(ctx context.Context, client *fivetran.Client, userId string) (users.UserConnectionMembershipsListResponse, error) {
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

/*          */
type UserGroupMemberships struct {
    UserId      types.String `tfsdk:"user_id"`
    Group       types.Set    `tfsdk:"group"`
}

func (d *UserGroupMemberships) ReadFromResponse(ctx context.Context, resp users.UserGroupMembershipsListResponse) {
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

func (d *UserGroupMemberships) ReadFromSource(ctx context.Context, client *fivetran.Client, userId string) (users.UserGroupMembershipsListResponse, error) {
    var respNextCursor string
    var listResponse users.UserGroupMembershipsListResponse
    limit := 1000

    svc := client.NewUserGroupMembershipsList()
    svc.UserId(userId)

    for {
        var err error
        var tmpResp users.UserGroupMembershipsListResponse

        if respNextCursor == "" {
            tmpResp, err = svc.Limit(limit).Do(ctx)
        }

        if respNextCursor != "" {
            tmpResp, err = svc.Limit(limit).Cursor(respNextCursor).Do(ctx)
        }
        
        if err != nil {
            listResponse = users.UserGroupMembershipsListResponse{}
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