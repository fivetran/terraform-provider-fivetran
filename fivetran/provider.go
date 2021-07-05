package fivetran

import (
	"context"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var limit = 1000 // REST API response objects limit per HTTP request

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key":    {Type: schema.TypeString, Required: true, DefaultFunc: schema.EnvDefaultFunc("FIVETRAN_APIKEY", nil)},
			"api_secret": {Type: schema.TypeString, Required: true, Sensitive: true, DefaultFunc: schema.EnvDefaultFunc("FIVETRAN_APISECRET", nil)},
		},
		ResourcesMap: map[string]*schema.Resource{
			"fivetran_user":  resourceUser(),
			"fivetran_group": resourceGroup(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"fivetran_user":        dataSourceUser(),
			"fivetran_users":       dataSourceUsers(),
			"fivetran_group":       dataSourceGroup(),
			"fivetran_group_users": dataSourceGroupUsers(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return fivetran.New(d.Get("api_key").(string), d.Get("api_secret").(string)), diag.Diagnostics{}
}
