package tests

import (
	"context"
	"os"
	"testing"

	gofivetran "github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testProviders map[string]*schema.Provider

func Client() *gofivetran.Client {
	return testProviders["fivetran-provider"].Meta().(*gofivetran.Client)
}

func init() {
	os.Setenv("FIVETRAN_APIKEY", "_moonbeam_acc_accountworthy_api_key")
	os.Setenv("FIVETRAN_APISECRET", "_moonbeam_acc_accountworthy_api_secret")
	provider := fivetran.Provider()
	provider.ConfigureContextFunc = providerConfigure
	testProviders = map[string]*schema.Provider{
		"fivetran-provider": provider,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	fivetranClient := gofivetran.New(d.Get("api_key").(string), d.Get("api_secret").(string))
	//Don't forget fo change version
	fivetranClient.BaseURL("http://localhost:8001/v1")
	return fivetranClient, diag.Diagnostics{}
}

func GetResource(t *testing.T, s *terraform.State, resourceName string) *terraform.ResourceState {
	// retrieve the resource by name from state
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		 t.Fatalf("Not found: %s", resourceName)
	}

	if rs.Primary.ID == "" {
		t.Fatalf(resourceName + " ID is not set")
	}

	return rs; 
}
