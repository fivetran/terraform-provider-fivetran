package model

import (
	"context"

	"github.com/fivetran/go-fivetran/proxy"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProxyAgents struct {
	Items types.Set `tfsdk:"items"`
}

func (d *ProxyAgents) ReadFromResponse(ctx context.Context, resp proxy.ProxyListResponse) {
	elementType := map[string]attr.Type{
		"id":           types.StringType,
		"registred_at": types.StringType,
		"group_region": types.StringType,
		"token":        types.StringType,
		"salt":         types.StringType,
		"created_by":   types.StringType,
		"display_name": types.StringType,
	}

	if resp.Data.Items == nil {
		d.Items = types.SetNull(types.ObjectType{AttrTypes: elementType})
	}

	items := []attr.Value{}

	for _, v := range resp.Data.Items {
		item := map[string]attr.Value{}
		item["id"] = types.StringValue(v.Id)
		item["registred_at"] = types.StringValue(v.RegisteredAt)
		item["group_region"] = types.StringValue(v.Region)
		item["token"] = types.StringValue(v.Token)
		item["salt"] = types.StringValue(v.Salt)
		item["created_by"] = types.StringValue(v.CreatedBy)
		item["display_name"] = types.StringValue(v.DisplayName)
		objectValue, _ := types.ObjectValue(elementType, item)
		items = append(items, objectValue)
	}

	d.Items, _ = types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
}
