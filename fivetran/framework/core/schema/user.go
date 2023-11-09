package schema

import (
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"
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
