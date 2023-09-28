package fivetran

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceExternalLogging() *schema.Resource {
    return &schema.Resource{
        ReadContext: resourceExternalLoggingRead,
        Schema:      getExternalLoggingSchema(true),
    }
}