package fivetran

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceTeamConnectorMembership() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceTeamConnectorMembershipRead,
		Schema: 	 getTeamConnectorMembershipSchema(true),
	}
}