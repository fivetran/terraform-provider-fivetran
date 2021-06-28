package fivetran

import (
	"context"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("FIVETRAN_APIKEY", nil),
			},
			"api_secret": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("FIVETRAN_APISECRET", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"fivetran_user": resourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"fivetran_user": dataSourceUser(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	apiKey := d.Get("api_key").(string)
	apiSecret := d.Get("api_secret").(string)

	c := fivetran.New(apiKey, apiSecret)

	return c, diags
}
