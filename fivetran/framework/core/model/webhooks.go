package model

import (
    "context"

    "github.com/fivetran/go-fivetran/webhooks"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type Webhooks struct {
    Webhooks   types.Set `tfsdk:"webhooks"`
}

func (d *Webhooks) ReadFromResponse(ctx context.Context, resp webhooks.WebhookListResponse) {
    webhookElementType := map[string]attr.Type{
        "id":           types.StringType,
        "type":         types.StringType,
        "group_id":     types.StringType,
        "url":          types.StringType,
        "active":       types.BoolType,
        "created_by":   types.StringType,
        "created_at":   types.StringType,
        "secret":       types.StringType,
        "events":       types.SetType{ElemType: types.StringType},
        "run_tests":    types.BoolType,
    }

    if resp.Data.Items == nil {
        d.Webhooks = types.SetNull(types.ObjectType{AttrTypes: webhookElementType})
    }

    items := []attr.Value{}
    
    for _, v := range resp.Data.Items {
        webhookItem := map[string]attr.Value{}
        webhookItem["id"] = types.StringValue(v.Id)
        webhookItem["type"] = types.StringValue(v.Type)
        webhookItem["group_id"] = types.StringValue(v.GroupId)
        webhookItem["url"] = types.StringValue(v.Url)
        webhookItem["events"], _ = types.SetValueFrom(ctx, types.StringType, v.Events)
        webhookItem["active"] = types.BoolValue(v.Active)
        webhookItem["secret"] = types.StringValue(v.Secret)
        webhookItem["created_at"] = types.StringValue(v.CreatedAt)
        webhookItem["created_by"] = types.StringValue(v.CreatedBy)
        webhookItem["run_tests"] = types.BoolValue(false)

        objectValue, _ := types.ObjectValue(webhookElementType, webhookItem)
        items = append(items, objectValue)
    }


    d.Webhooks, _ = types.SetValue(types.ObjectType{AttrTypes: webhookElementType}, items)
}