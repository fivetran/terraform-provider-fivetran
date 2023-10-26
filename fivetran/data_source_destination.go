package fivetran

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDestination() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceDestinationRead,
		Schema: 	 getDestinationSchema(true),
	}
}