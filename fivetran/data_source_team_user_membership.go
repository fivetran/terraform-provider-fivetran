package fivetran

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceTeamUserMembership() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceTeamUserMembershipRead,
		Schema: 	 getTeamUserMembershipSchema(true),
	}
}