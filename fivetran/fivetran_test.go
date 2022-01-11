package fivetran_test

import (
	"context"
	"log"
	"testing"

	gofivetran "github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testProviders map[string]*schema.Provider
var PredefinedGroupId string = "century_leveled"
var PredefinedUserId string = "endeavor_lock"
var providerFactory = make(map[string]func() (*schema.Provider, error))

func Client() *gofivetran.Client {
	return testProviders["fivetran-provider"].Meta().(*gofivetran.Client)
}

func init() {
	provider := fivetran.Provider()
	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		fivetranClient := gofivetran.New(d.Get("api_key").(string), d.Get("api_secret").(string))
		//Don't forget fo change version
		fivetranClient.BaseURL("https://api.fivetran.com/v1")
		return fivetranClient, diag.Diagnostics{}
	}

	testProviders = map[string]*schema.Provider{
		"fivetran-provider": provider,
	}

	for key, element := range testProviders {
		providerFactory[key] = func() (*schema.Provider, error) {
			return element, nil
		}
	}
	cleanupAccount()
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

	return rs
}

func cleanupAccount() {
	cleanupUsers()
	cleanupDestinations()
	cleanupGroups()
}

func cleanupUsers() {
	users, err := Client().NewUsersList().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, user := range users.Data.Items {
		if user.ID != PredefinedUserId {
			_, err := Client().NewUserDelete().UserID(user.ID).Do(context.Background())
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func cleanupDestinations() {
	groups, err := Client().NewGroupsList().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, group := range groups.Data.Items {
		_, err := Client().NewDestinationDelete().DestinationID(group.ID).Do(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func cleanupGroups() {
	groups, err := Client().NewGroupsList().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, group := range groups.Data.Items {
		cleanupConnectors(group.ID)
		if group.ID != PredefinedGroupId {
			_, err := Client().NewGroupDelete().GroupID(group.ID).Do(context.Background())
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func cleanupConnectors(groupId string) {
	connectors, err := Client().NewGroupListConnectors().GroupID(groupId).Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, connector := range connectors.Data.Items {
		_, err := Client().NewConnectorDelete().ConnectorID(connector.ID).Do(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}
}
