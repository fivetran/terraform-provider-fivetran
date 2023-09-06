package fivetran

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceDbtProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceDbtProjectRead,
		Schema:      getDbtProjectSchema(true),
	}
}
