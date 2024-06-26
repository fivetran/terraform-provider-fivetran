package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TeamResource() resourceSchema.Schema {
	return resourceSchema.Schema {
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the team within your account.",
			},
			"name": resourceSchema.StringAttribute{
				Required:    true,
				Description: "The name of the team within your account.",
			},
			"description": resourceSchema.StringAttribute{
				Optional:	 true,
				Description: "The description of the team within your account.",
			},
			"role": resourceSchema.StringAttribute{
				Required:    true,
				Description: "The account role of the team.",
			},
		},
	}
}

func TeamDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the team within your account.",
			},
			"name": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The name of the team within your account.",
			},
			"description": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The description of the team within your account.",
			},
			"role": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The account role of the team.",
			},
		},
	}
}

func TeamsDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "id": datasourceSchema.StringAttribute{
                Optional:      true,
                Description:   "The ID of this resource.",
            },
        },
		Blocks: map[string]datasourceSchema.Block{
			"teams": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Required:    true,
							Description: "The unique identifier for the team within your account.",
						},
						"name": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The name of the team within your account.",
						},
						"description": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The description of the team within your account.",
						},
						"role": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The account role of the team.",
						},
					},
				},
			},
		},
	}
}