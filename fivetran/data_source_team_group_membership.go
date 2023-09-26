package fivetran

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceTeamGroupMembership() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceTeamGroupMembershipRead,
		Schema: 	 getTeamGroupMembershipSchema(true),
	}
}