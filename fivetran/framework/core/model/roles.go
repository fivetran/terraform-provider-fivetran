package model

import (
    "context"

    "github.com/fivetran/go-fivetran/roles"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type Roles struct {
    Id       types.String `tfsdk:"id"` 
    Roles    types.Set    `tfsdk:"roles"`
}

func (d *Roles) ReadFromResponse(ctx context.Context, resp roles.RolesListResponse) {
    roleElementType := map[string]attr.Type{
        "name":                  types.StringType,
        "description":           types.StringType,
        "is_custom":             types.BoolType,
        "scope":                 types.SetType{ElemType: types.StringType},
        "is_deprecated":         types.BoolType,
        "replacement_role_name": types.StringType,
    }

    if resp.Data.Items == nil {
        d.Roles = types.SetNull(types.ObjectType{AttrTypes: roleElementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["name"] = types.StringValue(v.Name)
        item["description"] = types.StringValue(v.Description)
        item["is_custom"] = types.BoolValue(*v.IsCustom)
        item["scope"], _ = types.SetValueFrom(ctx, types.StringType, v.Scope)
        item["is_deprecated"] = types.BoolValue(*v.IsDeprecated)
        item["replacement_role_name"] = types.StringValue(v.ReplacementRoleName)

        objectValue, _ := types.ObjectValue(roleElementType, item)
        items = append(items, objectValue)
    }

    d.Id = types.StringValue("0")
    d.Roles, _ = types.SetValue(types.ObjectType{AttrTypes: roleElementType}, items)
}