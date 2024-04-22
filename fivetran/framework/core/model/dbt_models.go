package model

import (
	"context"

	"github.com/fivetran/go-fivetran/dbt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type DbtModels struct {
	Id        types.String `tfsdk:"id"`
	ProjectId types.String `tfsdk:"project_id"`
	Models    types.Set    `tfsdk:"models"`
}

func (d *DbtModels) ReadFromResponse(ctx context.Context, resp dbt.DbtModelsListResponse) {
	d.Models = GetModelsSetFromResponse(resp)
	d.Id = d.ProjectId

}

func GetModelsSetFromResponse(resp dbt.DbtModelsListResponse) basetypes.SetValue {
	elementType := ModelElementType()

	if resp.Data.Items == nil {
		return types.SetNull(types.ObjectType{AttrTypes: elementType})
	}

	items := []attr.Value{}

	for _, v := range resp.Data.Items {
		item := map[string]attr.Value{}
		item["id"] = types.StringValue(v.ID)
		item["model_name"] = types.StringValue(v.ModelName)
		item["scheduled"] = types.BoolValue(v.Scheduled)

		objectValue, _ := types.ObjectValue(elementType, item)
		items = append(items, objectValue)
	}
	result, _ := types.SetValue(types.ObjectType{AttrTypes: elementType}, items)
	return result
}

func ModelElementType() map[string]attr.Type {
	return map[string]attr.Type{
		"id":         types.StringType,
		"model_name": types.StringType,
		"scheduled":  types.BoolType,
	}
}
