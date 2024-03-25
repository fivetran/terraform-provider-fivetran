package schema

import (
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func GroupUsersDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				Required:    true,
				Description: "The unique identifier for the group within the Fivetran system. Data-source will represent a set of users who has membership in this group.",
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"users": datasourceSchema.SetNestedBlock{
                NestedObject: datasourceSchema.NestedBlockObject{
                    Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the user within the account.",
						},
						"email": datasourceSchema.StringAttribute{
							Required:    true,
							Description: "The email address that the user has associated with their user profile.",
						},
						"given_name": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The first name of the user.",
						},
						"family_name": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The last name of the user.",
						},
						"verified": datasourceSchema.BoolAttribute{
							Computed:    true,
							Description: "The field indicates whether the user has verified their email address in the account creation process.",
						},
						"invited": datasourceSchema.BoolAttribute{
							Computed:    true,
							Description: "The field indicates whether the user has verified their email address in the account creation process.",
						},
						"picture": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The user's avatar as a URL link (for example, 'http://mycompany.com/avatars/john_white.png') or base64 data URI (for example, 'data:image/png;base64,aHR0cDovL215Y29tcGFueS5jb20vYXZhdGFycy9qb2huX3doaXRlLnBuZw==')",
						},
						"phone": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The phone number of the user.",
						},
						"role": datasourceSchema.StringAttribute{
							Required:    true,
							Description: "The group role that you would like to assign this new user to. Supported group roles: ‘Destination Administrator‘, ‘Destination Reviewer‘, ‘Destination Analyst‘, ‘Connector Creator‘, or a custom destination role",
						},
						"logged_in_at": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The last time that the user has logged into their Fivetran account.",
						},
						"created_at": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The timestamp that the user created their Fivetran account",
						},
					},
				},
			},
		},
	}
}

func GroupUsersResource() resourceSchema.Schema {
	return resourceSchema.Schema {
		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for the resource.",
			},
			"group_id": resourceSchema.StringAttribute{
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Description: "The unique identifier for the Group within the Fivetran system.",
			},
			"last_updated": resourceSchema.StringAttribute{
				Computed:    true,
				Description: "",
			},
		},
		Blocks:  map[string]resourceSchema.Block{
			"user": resourceSchema.SetNestedBlock{
                NestedObject: resourceSchema.NestedBlockObject{
                    Attributes: map[string]resourceSchema.Attribute{
						"id": resourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the user within the account.",
						},
						"email": resourceSchema.StringAttribute{
							Required:    true,
							Description: "The email address that the user has associated with their user profile.",
						},
						"role": resourceSchema.StringAttribute{
							Required:    true,
							Description: "The group role that you would like to assign this new user to. Supported group roles: ‘Destination Administrator‘, ‘Destination Reviewer‘, ‘Destination Analyst‘, ‘Connector Creator‘, or a custom destination role",
						},
                    },
                },
            },
		},
	}
}