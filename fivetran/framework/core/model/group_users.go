package model

import (
    "context"

    "github.com/fivetran/go-fivetran/groups"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type GroupUsers struct {
	Id      types.String `tfsdk:"id"`
    Users   types.Set    `tfsdk:"users"`
}

func (d *GroupUsers) ReadFromResponse(ctx context.Context, resp groups.GroupListUsersResponse) {
    elementType := map[string]attr.Type{
        "id":           types.StringType,
        "email":        types.StringType,
        "given_name":  	types.StringType,
        "family_name":  types.StringType,
		"verified":     types.BoolType,
		"invited":      types.BoolType,
		"picture":      types.StringType,
		"phone":        types.StringType,
		"role":         types.StringType,
		"logged_in_at": types.StringType,
		"created_at":   types.StringType,

    }

    if resp.Data.Items == nil {
        d.Users = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    users := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        user := map[string]attr.Value{}
		user["id"] = types.StringValue(v.ID)
		user["email"] = types.StringValue(v.Email)
		user["given_name"] = types.StringValue(v.GivenName)
		user["family_name"] = types.StringValue(v.FamilyName)
		user["verified"] = types.BoolValue(*v.Verified)
		user["invited"] = types.BoolValue(*v.Invited)
		user["picture"] = types.StringValue(v.Picture)
		user["phone"] = types.StringValue(v.Phone)
		user["role"] = types.StringValue(v.Role)
		user["logged_in_at"] = types.StringValue(v.LoggedInAt.String())
		user["created_at"] = types.StringValue(v.CreatedAt.String())

        objectValue, _ := types.ObjectValue(elementType, user)
        users = append(users, objectValue)
    }

    d.Id = types.StringValue("0")
    d.Users, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, users)
}