package model

import (
	"context"

	"github.com/fivetran/go-fivetran/transformations"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type QuickstartPackage struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Version          types.String `tfsdk:"version"`
	ConnectorTypes   types.Set    `tfsdk:"connector_types"`
	OutputModelNames types.Set    `tfsdk:"output_model_names"`
	ConfigurableVars types.Map    `tfsdk:"configurable_vars"`
}

var configurableVarAttrTypes = map[string]attr.Type{
	"type":           types.StringType,
	"description":    types.StringType,
	"allowed_values": types.ListType{ElemType: types.StringType},
}

func (d *QuickstartPackage) ReadFromResponse(ctx context.Context, resp transformations.QuickstartPackageResponse) {
	d.Id = types.StringValue(resp.Data.Id)
	d.Name = types.StringValue(resp.Data.Name)
	d.Version = types.StringValue(resp.Data.Version)

	connectors := []attr.Value{}
	for _, connector := range resp.Data.ConnectorTypes {
		connectors = append(connectors, types.StringValue(connector))
	}
	d.ConnectorTypes, _ = types.SetValue(types.StringType, connectors)

	models := []attr.Value{}
	for _, model := range resp.Data.OutputModelNames {
		models = append(models, types.StringValue(model))
	}
	d.OutputModelNames, _ = types.SetValue(types.StringType, models)

	d.ConfigurableVars = configurableVarsFromResponse(resp.Data.ConfigurableVars)
}

func configurableVarsFromResponse(vars map[string]transformations.ConfigurableVarDefinition) types.Map {
	if len(vars) == 0 {
		return types.MapNull(types.ObjectType{AttrTypes: configurableVarAttrTypes})
	}
	items := map[string]attr.Value{}
	for name, v := range vars {
		allowedVals := []attr.Value{}
		for _, av := range v.AllowedValues {
			allowedVals = append(allowedVals, types.StringValue(av))
		}
		allowedList, _ := types.ListValue(types.StringType, allowedVals)
		obj, _ := types.ObjectValue(configurableVarAttrTypes, map[string]attr.Value{
			"type":           types.StringValue(v.Type),
			"description":    types.StringValue(v.Description),
			"allowed_values": allowedList,
		})
		items[name] = obj
	}
	result, _ := types.MapValue(types.ObjectType{AttrTypes: configurableVarAttrTypes}, items)
	return result
}