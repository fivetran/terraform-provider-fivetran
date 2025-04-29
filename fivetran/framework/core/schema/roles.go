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
							Description:   "Defines whether the role is standard or custom",
						},
						"scope": datasourceSchema.SetAttribute{
							Computed:      true,
							Description:   "Defines the list of resources the role manages. Supported values: ACCOUNT, DESTINATION, CONNECTOR, and TEAM",
							ElementType:   types.StringType,
						},
						"is_deprecated": datasourceSchema.BoolAttribute{
							Computed:      true,
							Description:   "Defines whether the role is deprecated",
						},
						"replacement_role_name": datasourceSchema.StringAttribute{
							Computed:      true,
							Description:   "The name of the new role replacing the deprecated role",
						},
					},
				},
			},
		},
	}
}