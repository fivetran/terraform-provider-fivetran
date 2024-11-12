package model

import (
    "context"

    localprocessingagent "github.com/fivetran/go-fivetran/hybrid_deployment_agent"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type LocalProcessingAgents struct {
    Items   types.Set `tfsdk:"items"`
}

func (d *LocalProcessingAgents) ReadFromResponse(ctx context.Context, resp localprocessingagent.HybridDeploymentAgentListResponse) {
    subSetElementType := map[string]attr.Type{
        "connection_id":   types.StringType,
        "schema":          types.StringType,
        "service":         types.StringType,
    }

    subSetAttrType := types.ObjectType{
        AttrTypes: subSetElementType,
    }

    elementType := map[string]attr.Type{
        "id":              types.StringType,
        "display_name":    types.StringType,
        "group_id":        types.StringType,
        "registered_at":   types.StringType,
        "usage":           types.SetType{ElemType: subSetAttrType},
    }

    if resp.Data.Items == nil {
        d.Items = types.SetNull(types.ObjectType{AttrTypes: elementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["id"] = types.StringValue(v.Id)
        item["display_name"] = types.StringValue(v.DisplayName)
        item["group_id"] = types.StringValue(v.GroupId)
        item["registered_at"] = types.StringValue(v.RegisteredAt)

        subItems := []attr.Value{}
        for _, sub := range v.Usage {
            subItem := map[string]attr.Value{}
            subItem["connection_id"] = types.StringValue(sub.ConnectionId)
            subItem["schema"] = types.StringValue(sub.Schema)
            subItem["service"] = types.StringValue(sub.Service)

            subObjectValue, _ := types.ObjectValue(subSetElementType, subItem)
            subItems = append(subItems, subObjectValue)
        }

        item["usage"], _ = types.SetValue(subSetAttrType, subItems)

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Items, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}