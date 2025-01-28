package model

import (
	"context"

	"github.com/fivetran/go-fivetran/transformations"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type QuickstartPackage struct {
	Id              	types.String `tfsdk:"id"`
	Name         		types.String `tfsdk:"name"`
	Version  			types.String `tfsdk:"version"`
	ConnectorTypes  	types.Set 	 `tfsdk:"connector_types"`
	OutputModelNames 	types.Set    `tfsdk:"output_model_names"`
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
	for _, connector := range resp.Data.OutputModelNames {
		models = append(models, types.StringValue(connector))
	}
	d.OutputModelNames, _ = types.SetValue(types.StringType, models)
}