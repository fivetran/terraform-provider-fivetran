package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)


func GroupResource() resourceSchema.Schema {
	return resourceSchema.Schema {
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:      true,
				Description:   "The unique identifier for the group within the Fivetran system.",
			},
			"name": resourceSchema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description:   "The name of the group within your account.",
			},
			"created_at": resourceSchema.StringAttribute{
				Computed:      true,
				Description:   "The timestamp of when the group was created in your account.",
			},
			"last_updated": resourceSchema.StringAttribute{
				Computed:      true,
				Description:   "The timestamp of when the group was updated in your account.",
			},
		},
	}
}

func GroupDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Required:      true,
				Description:   "The unique identifier for the group within the Fivetran system.",
			},
			"name": datasourceSchema.StringAttribute{
				Computed:      true,
				Description:   "The name of the group within your account.",
			},
			"created_at": datasourceSchema.StringAttribute{
				Computed:      true,
				Description:   "The timestamp of when the group was created in your account.",
			},
			"last_updated": resourceSchema.StringAttribute{
				Computed:      true,
				Description:   "The timestamp of when the group was updated in your account.",
			},
		},
	}
}

func GroupsDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"groups": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Computed:      true,
							Description:   "The unique identifier for the group within the Fivetran system.",
						},
						"name": datasourceSchema.StringAttribute{
							Computed:      true,
							Description:   "The name of the group within your account.",
						},
						"created_at": datasourceSchema.StringAttribute{
							Computed:      true,
							Description:   "The timestamp of when the group was created in your account.",
						},
						"last_updated": datasourceSchema.StringAttribute{
							Optional:      true,
							Description:   "The timestamp of when the group was updated in your account.",
						},
					},
				},
			},
		},
	}
}