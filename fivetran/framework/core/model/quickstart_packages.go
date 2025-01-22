package model

import (
	"context"

	"github.com/fivetran/go-fivetran/transformations"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type QuickstartPackages struct {
	Packages types.List `tfsdk:"packages"`
}

func (d *QuickstartPackages) ReadFromResponse(ctx context.Context, resp transformations.QuickstartPackagesListResponse) {
	elemTypeAttrs := map[string]attr.Type{
		"id":            		types.StringType,
		"name":      			types.StringType,
		"version":    			types.StringType,
        "connector_types":     	types.SetType{ElemType: types.StringType},
        "output_model_names":   types.SetType{ElemType: types.StringType},
	}

	if resp.Data.Items == nil {
		d.Packages = types.ListNull(types.ObjectType{AttrTypes: elemTypeAttrs})
	} else {
		items := []attr.Value{}
		for _, v := range resp.Data.Items {
			item := map[string]attr.Value{}
			item["id"] = types.StringValue(v.Id)
			item["name"] = types.StringValue(v.Name)
			item["version"] = types.StringValue(v.Version)

    		connectors := []attr.Value{}
    		for _, el := range v.ConnectorTypes {
        		connectors = append(connectors, types.StringValue(el))
    		}
    		
    		if len(connectors) > 0 {
        		item["connector_types"] = types.SetValueMust(types.StringType, connectors)
    		} else {
        		item["connector_types"] = types.SetNull(types.StringType)
    		}

    		models := []attr.Value{}
    		for _, el := range v.OutputModelNames {
        		models = append(models, types.StringValue(el))
    		}
    		
    		if len(models) > 0 {
        		item["output_model_names"] = types.SetValueMust(types.StringType, models)
    		} else {
        		item["output_model_names"] = types.SetNull(types.StringType)
    		}

			objectValue, _ := types.ObjectValue(elemTypeAttrs, item)
			items = append(items, objectValue)
		}

		d.Packages, _ = types.ListValue(types.ObjectType{AttrTypes: elemTypeAttrs}, items)
	}
}
