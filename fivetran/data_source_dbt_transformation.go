package fivetran

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceDbtTransformation() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceDbtTransformationRead,
		Schema:      getDbtTransformationSchema(true),
	}
}
