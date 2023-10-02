package fivetran

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceTeamRead,
		Schema: 	 getTeamSchema(true),
	}
}