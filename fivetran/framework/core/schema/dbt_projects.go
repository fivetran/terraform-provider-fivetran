package schema

import datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

func DbtProjectsSchema() datasourceSchema.Schema {
	return datasourceSchema.Schema{
		DeprecationMessage: "This resource is Deprecated, please follow the 1.5.0 migration guide to update the schema",
		Attributes: map[string]datasourceSchema.Attribute{
			"projects": datasourceSchema.ListNestedAttribute{
				Computed: true,
				NestedObject: datasourceSchema.NestedAttributeObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the dbt project within the Fivetran system.",
						},
						"group_id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The name of the group within your account related to the project.",
						},
						"created_at": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The timestamp of when the project was created in your account.",
						},
						"created_by_id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the User within the Fivetran system who created the DBT Project.",
						},
					},
				},
			},
		},
	}
}
