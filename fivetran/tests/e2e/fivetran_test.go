package e2e_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"math/rand"
	"runtime"
	"time"

	gofivetran "github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	// ! WARNING !
	// ! Before changing these values usure you're using BLANK ACCOUNT API KEY. All data from account will be deleted !
	PredefinedGroupId       = "harbour_choking"
	PredefinedUserId        = "buyer_warring"
	PredefinedUserGivenName = "Terraform"
	PredefinedGroupName     = "Warehouse"
	BqProjectId             = "dulcet-yew-246109"
)

var testProvioderFramework provider.Provider
var client *gofivetran.Client
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"fivetran-provider": func() (tfprotov6.ProviderServer, error) {
		return providerserver.NewProtocol6(testProvioderFramework)(), nil
	},
}

func init() {
	// uncomment for local testing
	// os.Setenv("FIVETRAN_API_URL", "https://api-staging.fivetran.com/v1")
	// os.Setenv("FIVETRAN_APIKEY", "apiKey")
	// os.Setenv("FIVETRAN_APISECRET", "apiSecret")
	// os.Setenv("TF_ACC", "True")

	var apiUrl, apiKey, apiSecret string
	valuesToLoad := map[string]*string{
		"FIVETRAN_API_URL":   &apiUrl,
		"FIVETRAN_APIKEY":    &apiKey,
		"FIVETRAN_APISECRET": &apiSecret,
	}

	for name, value := range valuesToLoad {
		*value = os.Getenv(name)
		if *value == "" {
			log.Fatalf("Environment variable %s is not set!\n", name)
		}
	}

	client = gofivetran.New(apiKey, apiSecret)
	client.BaseURL(apiUrl)

	testProvioderFramework = framework.FivetranProvider()

	if isPredefinedUserExist() && isPredefinedGroupExist() {
		cleanupAccount()
	} else {
		log.Fatalln("The predefined user doesn't belong to the Testing account. Make sure that credentials are using in the tests belong to the Testing account.")
	}
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
	cleanupWebhooks()
	cleanupTeams()
	cleanupConnections()
	cleanupProxyAgents()
	cleanupHybridDeploymentAgents()
	// cleanupPrivateLinks() 
	cleanupExternalLogging()
	cleanupDestinations()
	cleanupGroups()
}

func isPredefinedUserExist() bool {
	user, err := client.NewUserDetails().UserID(PredefinedUserId).Do(context.Background())
	if err != nil {
		return false
	}
	return user.Data.GivenName == PredefinedUserGivenName
}

func isPredefinedGroupExist() bool {
	group, err := client.NewGroupDetails().GroupID(PredefinedGroupId).Do(context.Background())
	if err != nil {
		return false
	}
	return group.Data.Name == PredefinedGroupName
}

func cleanupUsers() {
	list, err := client.NewUsersList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, item := range list.Data.Items {
		if item.ID != PredefinedUserId {
			_, err := client.NewUserDelete().UserID(item.ID).Do(context.Background())
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func cleanupDestinations() {
	list, err := client.NewDestinationsList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, item := range list.Data.Items {
		_, err := client.NewDestinationDelete().DestinationID(item.ID).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatalln(err)
		}
	}

	if list.Data.NextCursor != "" {
		cleanupDestinations()
	}
}

func cleanupExternalLogging() {
	list, err := client.NewExternalLoggingList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, item := range list.Data.Items {
		_, err := client.NewExternalLoggingDelete().ExternalLoggingId(item.Id).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatalln(err)
		}
	}

	if list.Data.NextCursor != "" {
		cleanupExternalLogging()
	}
}

func cleanupGroups() {
	list, err := client.NewGroupsList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, item := range list.Data.Items {
		if item.ID != PredefinedGroupId {
			_, err := client.NewGroupDelete().GroupID(item.ID).Do(context.Background())
			if err != nil && err.Error() != "status code: 404; expected: 200" {
				log.Fatalln(err)
			}
		}
	}
}

func cleanupConnections() {
	list, err := client.NewConnectionsList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, item := range list.Data.Items {
		_, err := client.NewConnectionDelete().ConnectionID(item.ID).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatalln(err)
		}
	}

	if list.Data.NextCursor != "" {
		cleanupConnections()
	}
}

func cleanupWebhooks() {
	list, err := client.NewWebhookList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, item := range list.Data.Items {
		_, err := client.NewWebhookDelete().WebhookId(item.Id).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatalln(err)
		}
	}

	if list.Data.NextCursor != "" {
		cleanupWebhooks()
	}
}

func cleanupTeams() {
	list, err := client.NewTeamsList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, item := range list.Data.Items {
		_, err := client.NewTeamsDelete().TeamId(item.Id).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatalln("cleanupTeams Delete")
			log.Fatalln(err)
		}
	}

	if list.Data.NextCursor != "" {
		cleanupTeams()
	}
}

func cleanupProxyAgents() {
	list, err := client.NewProxyList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, proxy := range list.Data.Items {
		_, err := client.NewProxyDelete().ProxyId(proxy.Id).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatalln("cleanupProxyAgents Delete")
			log.Fatalln(err)
		}
	}

	if list.Data.NextCursor != "" {
		cleanupProxyAgents()
	}
}

func cleanupPrivateLinks() {
	list, err := client.NewPrivateLinkList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, item := range list.Data.Items {
		_, err := client.NewPrivateLinkDelete().PrivateLinkId(item.Id).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatalln(err)
		}
	}

	if list.Data.NextCursor != "" {
		cleanupPrivateLinks()
	}
}

func cleanupHybridDeploymentAgents() {
	list, err := client.NewHybridDeploymentAgentList().Do(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	for _, item := range list.Data.Items {
		_, err := client.NewHybridDeploymentAgentDelete().AgentId(item.Id).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatalln(err)
		}
	}

	if list.Data.NextCursor != "" {
		cleanupHybridDeploymentAgents()
	}
}


func ComposeImportStateCheck(fs ...resource.ImportStateCheckFunc) resource.ImportStateCheckFunc {
	return func(s []*terraform.InstanceState) error {
		for i, f := range fs {
			if err := f(s); err != nil {
				return fmt.Errorf("check %d/%d error: %s", i+1, len(fs), err)
			}
		}

		return nil
	}
}

func CheckImportResourceAttr(resourceType, attributeName, value string) resource.ImportStateCheckFunc {
	_, file, line, _ := runtime.Caller(1)

	return func(s []*terraform.InstanceState) error {
		for _, v := range s {
			if v.Ephemeral.Type == resourceType {

				if attrVal, ok := v.Attributes[attributeName]; ok {
					if attrVal != value {
						return fmt.Errorf("For %s, '%s' attribute value is expected: '%s', got: '%s'. At %s:%d", v.Ephemeral.Type, attributeName, value, attrVal, file, line)
					}

					return nil
				} else {
					return fmt.Errorf("Attribute '%s' not found for %s. At %s:%d", attributeName, v.Ephemeral.Type, file, line)
				}
			}
		}

		return fmt.Errorf("Resource with type '%s' not found in imported state. At %s:%d", resourceType, file, line)
	}
}

func CheckImportResourceAttrSet(resourceType, attributeName string) resource.ImportStateCheckFunc {
	_, file, line, _ := runtime.Caller(1)

	return func(s []*terraform.InstanceState) error {
		for _, v := range s {
			if v.Ephemeral.Type == resourceType {

				if attrVal, ok := v.Attributes[attributeName]; ok {
					if attrVal != "" {
						return nil
					}

					return fmt.Errorf("For %s, '%s' attribute value is expected to be set, got: '%s'. At %s:%d", v.Ephemeral.Type, attributeName, attrVal, file, line)
				} else {
					return fmt.Errorf("Attribute '%s' not found for %s. At %s:%d", attributeName, v.Ephemeral.Type, file, line)
				}
			}
		}

		return fmt.Errorf("Resource with type '%s' not found in imported state. At %s:%d", resourceType, file, line)
	}
}

func CheckNoImportResourceAttr(resourceType, attributeName string) resource.ImportStateCheckFunc {
	_, file, line, _ := runtime.Caller(1)

	return func(s []*terraform.InstanceState) error {
		for _, v := range s {
			if v.Ephemeral.Type == resourceType {

				if attrVal, ok := v.Attributes[attributeName]; ok {
					if attrVal != "" {
						return fmt.Errorf("For %s, '%s' attribute is found while not expected. Got: '%s'. At %s:%d", v.Ephemeral.Type, attributeName, attrVal, file, line)
					}
				}

				return nil
			}
		}

		return fmt.Errorf("Resource with type '%s' not found in imported state. At %s:%d", resourceType, file, line)
	}
}