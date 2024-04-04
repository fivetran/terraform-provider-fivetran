package model

import (
    "context"

    "github.com/fivetran/go-fivetran/users"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type Users struct {
    Id       types.String `tfsdk:"id"` 
    Users    types.Set    `tfsdk:"users"`
}

func (d *Users) ReadFromResponse(ctx context.Context, resp users.UsersListResponse) {
    elementType := map[string]attr.Type{
        "id":           types.StringType,
        "email":        types.StringType,
        "given_name":   types.StringType,
        "family_name":  types.StringType,
        "verified":     types.BoolType,
        "invited":      types.BoolType,
        "picture":      types.StringType,
        "phone":        types.StringType,
        "logged_in_at": types.StringType,
        "created_at":   types.StringType,
    }

    if resp.Data.Items == nil {
        d.Users = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["id"] = types.StringValue(v.ID)
        item["email"] = types.StringValue(v.Email)
        item["given_name"] = types.StringValue(v.GivenName)
        item["family_name"] = types.StringValue(v.FamilyName)
        item["verified"] = types.BoolValue(*v.Verified)
        item["invited"] = types.BoolValue(*v.Invited)
        item["picture"] = types.StringValue(v.Picture)
        item["phone"] = types.StringValue(v.Phone)
        item["logged_in_at"] = types.StringValue(v.LoggedInAt.String())
        item["created_at"] = types.StringValue(v.CreatedAt.String())

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Id = types.StringValue("0")
    d.Users, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}