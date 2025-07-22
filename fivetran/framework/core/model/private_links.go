package model

import (
    "context"

    "github.com/fivetran/go-fivetran/private_link"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type PrivateLinks struct {
    Items   types.Set `tfsdk:"items"`
}

func (d *PrivateLinks) ReadFromResponse(ctx context.Context, resp privatelink.PrivateLinkListResponse) {
    elementType := map[string]attr.Type{
        "id":              types.StringType,
        "region":          types.StringType,
        "name":            types.StringType,
        "service":         types.StringType,
        "cloud_provider":  types.StringType,
        "state":           types.StringType,
        "state_summary":   types.StringType,
        "created_by":      types.StringType,
        "created_at":      types.StringType,
        "host":            types.StringType,
    }

    if resp.Data.Items == nil {
        d.Items = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["id"] = types.StringValue(v.Id)
        item["region"] = types.StringValue(v.Region)
        item["name"] = types.StringValue(v.Name)
        item["service"] = types.StringValue(v.Service)
        item["cloud_provider"] = types.StringValue(v.CloudProvider)
        item["state"] = types.StringValue(v.State)
        item["state_summary"] = types.StringValue(v.StateSummary)
        item["created_at"] = types.StringValue(v.CreatedAt)
        item["created_by"] = types.StringValue(v.CreatedBy)
        item["host"] = types.StringValue(v.Host)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Items, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}