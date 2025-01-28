package model

import (
	"context"

	"github.com/fivetran/go-fivetran/transformations"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TransformationProjects struct {
	Projects types.List `tfsdk:"projects"`
}

func (d *TransformationProjects) ReadFromResponse(ctx context.Context, resp transformations.TransformationProjectsListResponse) {
	elemTypeAttrs := map[string]attr.Type{
		"id":            types.StringType,
		"type":          types.StringType,
		"group_id":      types.StringType,
		"created_at":    types.StringType,
		"created_by_id": types.StringType,
	}

	if resp.Data.Items == nil {
		d.Projects = types.ListNull(types.ObjectType{AttrTypes: elemTypeAttrs})
	} else {
		items := []attr.Value{}
		for _, v := range resp.Data.Items {
			item := map[string]attr.Value{}
			item["id"] = types.StringValue(v.Id)
			item["type"] = types.StringValue(v.ProjectType)
			item["group_id"] = types.StringValue(v.GroupId)
			item["created_at"] = types.StringValue(v.CreatedAt)
			item["created_by_id"] = types.StringValue(v.CreatedById)
			objectValue, _ := types.ObjectValue(elemTypeAttrs, item)
			items = append(items, objectValue)
		}
		d.Projects, _ = types.ListValue(types.ObjectType{AttrTypes: elemTypeAttrs}, items)
	}
}