package schema

import (
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)
func ConnectorsMetadataDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"sources": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the connector within the Fivetran system",
						},
						"name": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The connector service name within the Fivetran system.",
						},
						"type": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The connector service type within the Fivetran system.",
						},
						"description": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The description characterizing the purpose of the connector.",
						},
						"icon_url": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The icon resource URL.",
						},
						"link_to_docs": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The link to the connector documentation.",
						},
						"link_to_erd": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The link to the connector ERD (entity–relationship diagram).",
						},
					},
				},
			},
		},
	}
}