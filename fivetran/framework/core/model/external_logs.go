package model

import (
    "context"

    "github.com/fivetran/go-fivetran/external_logging"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type ExternalLogs struct {
    Id              types.String `tfsdk:"id"` 
    ExternalLogs    types.Set    `tfsdk:"logs"`
}

func (d *ExternalLogs) ReadFromResponse(ctx context.Context, resp externallogging.ExternalLoggingListResponse) {
    elementAttrType := map[string]attr.Type{
        "id":      types.StringType,
        "service": types.StringType,
        "enabled": types.BoolType,
    }

    if resp.Data.Items == nil {
        d.ExternalLogs = types.SetNull(types.ObjectType{AttrTypes: elementAttrType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        item := map[string]attr.Value{}
        item["id"] = types.StringValue(v.Id)
        item["service"] = types.StringValue(v.Service)
        item["enabled"] = types.BoolValue(v.Enabled)

        objectValue, _ := types.ObjectValue(elementAttrType, item)
        items = append(items, objectValue)
    }

    d.Id = types.StringValue("0")
    d.ExternalLogs, _ = types.SetValue(types.ObjectType{AttrTypes: elementAttrType}, items)
}