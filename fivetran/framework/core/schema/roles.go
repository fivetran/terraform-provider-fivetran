package schema

import (
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

func RolesDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"roles": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"name": datasourceSchema.StringAttribute{
							Computed:      true,
							Description:   "The role name",
						},
						"description": datasourceSchema.StringAttribute{
							Computed:      true,
							Description:   "The role description",
						},
						"is_custom": datasourceSchema.BoolAttribute{
							Computed:      true,
							Description:   "TypeBool",
						},
						"scope": datasourceSchema.SetAttribute{
							Computed:      true,
							Description:   "Defines the list of resources the role manages. Supported values: ACCOUNT, DESTINATION, CONNECTOR, and TEAM",
							ElementType:   types.StringType,
						},
					},
				},
			},
		},
	}
}