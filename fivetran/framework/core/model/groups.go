package model

import (
    "context"

    "github.com/fivetran/go-fivetran/groups"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type Groups struct {
    Id       types.String `tfsdk:"id"` 
    Groups   types.Set    `tfsdk:"groups"`
}

func (d *Groups) ReadFromResponse(ctx context.Context, resp groups.GroupsListResponse) {
    elementType := map[string]attr.Type{
        "id":           types.StringType,
        "name":         types.StringType,
        "created_at":   types.StringType,
        "last_updated": types.StringType,
    }

    if resp.Data.Items == nil {
        d.Groups = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["id"] = types.StringValue(v.ID)
        item["name"] = types.StringValue(v.Name)
        item["created_at"] = types.StringValue(v.CreatedAt.String())
        item["last_updated"] = types.StringNull()

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Id = types.StringValue("0")
    d.Groups, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}