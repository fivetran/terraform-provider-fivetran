package schema

import "github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core"

func GroupSshKey() core.Schema {
	return core.Schema{
		Fields: map[string]core.SchemaField{
			"id": {
				IsId:        true,
				ValueType:   core.String,
				Description: "The unique identifier for the group within the Fivetran system.",
			},
			"public_key": {
				IsId:        false,
				ValueType:   core.String,
				Description: "Public key from SSH key pair associated with the group.",
			},
		},
	}
}
