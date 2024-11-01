package model

import (
    "context"

    "github.com/fivetran/go-fivetran/hybrid_deployment_agent"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type HybridDeploymentAgents struct {
    Items   types.Set `tfsdk:"items"`
}

func (d *HybridDeploymentAgents) ReadFromResponse(ctx context.Context, resp hybriddeploymentagent.HybridDeploymentAgentListResponse) {
    elementType := map[string]attr.Type{
        "id":              types.StringType,
        "display_name":    types.StringType,
        "group_id":        types.StringType,
        "registered_at":   types.StringType,
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

        objectValue, _ := types.ObjectValue(elementType, item)
        items = append(items, objectValue)
    }

    d.Items, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}