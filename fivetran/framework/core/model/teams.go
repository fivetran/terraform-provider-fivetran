package model

import (
    "context"

    "github.com/fivetran/go-fivetran/teams"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type Teams struct {
    Teams   types.Set `tfsdk:"teams"`
}

func (d *Teams) ReadFromResponse(ctx context.Context, resp teams.TeamsListResponse) {
    elementType := map[string]attr.Type{
        "id":           types.StringType,
        "name":         types.StringType,
        "description":  types.StringType,
        "role":         types.StringType,
    }

    if resp.Data.Items == nil {
        d.Teams = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["id"] = types.StringValue(v.Id)
        item["name"] = types.StringValue(v.Name)
        item["description"] = types.StringValue(v.Description)
        item["role"] = types.StringValue(v.Role)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }


    d.Teams, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}