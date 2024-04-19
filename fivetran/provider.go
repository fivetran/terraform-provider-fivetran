package fivetran

import (
	"context"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var limit = 1000 // REST API response objects limit per HTTP request

func Provider() *schema.Provider {
	var resourceMap = map[string]*schema.Resource{
		"fivetran_dbt_transformation": resourceDbtTransformation(),
		"fivetran_dbt_project":        resourceDbtProject(),
	}

	var dataSourceMap = map[string]*schema.Resource{
		"fivetran_dbt_transformation": dataSourceDbtTransformation(),
		"fivetran_dbt_project":        dataSourceDbtProject(),
		"fivetran_dbt_projects":       dataSourceDbtProjects(),
	}

	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key":    {Type: schema.TypeString, Optional: true},
			"api_secret": {Type: schema.TypeString, Optional: true, Sensitive: true},
			"api_url":    {Type: schema.TypeString, Optional: true},
		},
		ResourcesMap:         resourceMap,
		DataSourcesMap:       dataSourceMap,
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	if d.Get("api_key") == "" {
		apiKey, _ := schema.EnvDefaultFunc("FIVETRAN_APIKEY", nil)()
		d.Set("api_key", apiKey)
	}
	if d.Get("api_secret") == "" {
		apiSecret, _ := schema.EnvDefaultFunc("FIVETRAN_APISECRET", nil)()
		d.Set("api_secret", apiSecret)
	}
	if d.Get("api_url") == "" {
		apiUrl, _ := schema.EnvDefaultFunc("FIVETRAN_APIURL", nil)()
		d.Set("api_url", apiUrl)
	}

	fivetranClient := fivetran.New(d.Get("api_key").(string), d.Get("api_secret").(string))
	if d.Get("api_url") != "" {
		fivetranClient.BaseURL(d.Get("api_url").(string))
	}

	fivetranClient.CustomUserAgent("terraform-provider-fivetran/" + framework.Version)
	return fivetranClient, diag.Diagnostics{}
}
