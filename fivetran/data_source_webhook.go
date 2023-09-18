package fivetran

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceWebhook() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceWebhookRead,
		Schema: 	 getWebhookSchema(true),
	}
}