package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func User() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the user within the Fivetran system.",
			},
			"email": {
				Required:    true,
				ForceNew:    true,
				ValueType:   core.String,
				Description: "The email address that the user has associated with their user profile.",
			},
			"given_name": {
				Required:    true,
				ValueType:   core.String,
				Description: "The first name of the user.",
			},
			"family_name": {
				Required:    true,
				ValueType:   core.String,
				Description: "The last name of the user.",
			},
			"picture": {
				ValueType:   core.String,
				Description: "The user's avatar as a URL link (for example, 'http://mycompany.com/avatars/john_white.png') or base64 data URI (for example, 'data:image/png;base64,aHR0cDovL215Y29tcGFueS5jb20vYXZhdGFycy9qb2huX3doaXRlLnBuZw==')",
			},
			"phone": {
				ValueType:   core.String,
				Description: "The phone number of the user.",
			},
			"role": {
				ValueType:   core.String,
				Description: "The role that you would like to assign to the user.",
			},
			"invited": {
				ValueType:   core.Boolean,
				Description: "The field indicates whether the user has been invited to your account.",
			},
			"verified": {
				ValueType:   core.Boolean,
				Description: "The field indicates whether the user has verified their email address in the account creation process.",
			},
			"logged_in_at": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The last time that the user has logged into their Fivetran account.",
			},
			"created_at": {
				Readonly:    true,
				ValueType:   core.String,
				Description: "The timestamp that the user created their Fivetran account.",
			},
		},
	}
}

func UsersDatasource() datasourceSchema.Schema {
	return datasourceSchema.Schema {
        Attributes: map[string]datasourceSchema.Attribute{
            "id": datasourceSchema.StringAttribute{
                Optional:      true,
                Description:   "The ID of this resource.",
            },
        },
		Blocks: map[string]datasourceSchema.Block{
			"users": datasourceSchema.SetNestedBlock{
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"id": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier for the user within your account.",
						},
						"email": datasourceSchema.StringAttribute{
							Computed:    true,
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
							Description: "The field indicates whether the user has been invited to your account.",
						},
						"picture": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The user's avatar as a URL link (for example, 'http://mycompany.com/avatars/john_white.png') or base64 data URI (for example, 'data:image/png;base64,aHR0cDovL215Y29tcGFueS5jb20vYXZhdGFycy9qb2huX3doaXRlLnBuZw==')",
						},
						"phone": datasourceSchema.StringAttribute{
							Computed:    true,
							Description: "The phone number of the user.",
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