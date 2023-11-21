package schema

import "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"

func GroupServiceAccount() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the group within the Fivetran system.",
			},
			"service_account": {
				IsId:        false,
				ValueType:   core.String,
				Description: "Fivetran service account associated with the group.",
			},
		},
	}
}
