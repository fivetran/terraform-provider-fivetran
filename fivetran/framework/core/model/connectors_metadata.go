package model

import (
    "context"

    "github.com/fivetran/go-fivetran/metadata"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type ConnectorsMetadata struct {
    Id       types.String `tfsdk:"id"` 
    Sources  types.Set    `tfsdk:"sources"`
}

func (d *ConnectorsMetadata) ReadFromResponse(ctx context.Context, resp metadata.ConnectorMetadataListResponse) {
    elementType := map[string]attr.Type{
        "id":            types.StringType,
        "name":          types.StringType,
        "type":          types.StringType,
        "description":   types.StringType,
        "icon_url":      types.StringType,
        "link_to_docs":  types.StringType,
        "link_to_erd":   types.StringType,
    }

    if resp.Data.Items == nil {
        d.Sources = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["id"] = types.StringValue(v.ID)
        item["name"] = types.StringValue(v.Name)
        item["type"] = types.StringValue(v.Type)
        item["description"] = types.StringValue(v.Description)
        item["icon_url"] = types.StringValue(v.IconURL)
        item["link_to_docs"] = types.StringValue(v.LinkToDocs)
        item["link_to_erd"] = types.StringValue(v.LinkToErd)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Id = types.StringValue("0")
    d.Sources, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}