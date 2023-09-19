package fivetran_test

import (
	"context"
	"log"
	"os"
	"testing"

	gofivetran "github.com/fivetran/go-fivetran"
	"github.com/fivetran/terraform-provider-fivetran/fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

var client *gofivetran.Client
var testProviders map[string]*schema.Provider
var providerFactory = make(map[string]func() (*schema.Provider, error))

func init() {
	// uncomment for local testing
	// os.Setenv("FIVETRAN_API_URL", "https://api-staging.fivetran.com/v1")
	// os.Setenv("FIVETRAN_APIKEY", "apikey")
	// os.Setenv("FIVETRAN_APISECRET", "apisecret")
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

	provider := fivetran.Provider()
	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return client, diag.Diagnostics{}
	}

	testProviders = map[string]*schema.Provider{
		"fivetran-provider": provider,
	}

	for key, element := range testProviders {
		providerFactory[key] = func() (*schema.Provider, error) {
			return element, nil
		}
	}

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
	cleanupExternalLogging()
	cleanupDestinations()
	cleanupDbtProjects()
	cleanupGroups()
	cleanupWebhooks()
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
	users, err := client.NewUsersList().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, user := range users.Data.Items {
		if user.ID != PredefinedUserId {
			_, err := client.NewUserDelete().UserID(user.ID).Do(context.Background())
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func cleanupDestinations() {
	groups, err := client.NewGroupsList().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, group := range groups.Data.Items {
		_, err := client.NewDestinationDelete().DestinationID(group.ID).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatal(err)
		}
	}
}

func cleanupExternalLogging() {
	groups, err := client.NewGroupsList().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, group := range groups.Data.Items {
		_, err := client.NewExternalLoggingDelete().ExternalLoggingId(group.ID).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatal(err)
		}
	}
}

func cleanupDbtProjects() {
	projects, err := client.NewDbtProjectsList().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, project := range projects.Data.Items {
		cleanupDbtTransformations(project.ID, "")
		_, err := client.NewDbtProjectDelete().DbtProjectID(project.ID).Do(context.Background())
		if err != nil && err.Error() != "status code: 404; expected: 200" {
			log.Fatal(err)
		}
	}
	if projects.Data.NextCursor != "" {
		cleanupDbtProjects()
	}
}

func cleanupDbtTransformations(projectId, nextCursor string) {
	svc := client.NewDbtModelsList().ProjectId(projectId)

	if nextCursor != "" {
		svc.Cursor(nextCursor)
	}

	models, err := svc.Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, model := range models.Data.Items {
		if model.Scheduled {
			_, err := client.NewDbtTransformationDeleteService().TransformationId(model.ID).Do(context.Background())
			if err != nil && err.Error() != "status code: 404; expected: 200" {
				log.Fatal(err)
			}
		}
	}

	if models.Data.NextCursor != "" {
		cleanupDbtTransformations(projectId, models.Data.NextCursor)
	}
}

func cleanupGroups() {
	groups, err := client.NewGroupsList().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, group := range groups.Data.Items {
		cleanupConnectors(group.ID)
		if group.ID != PredefinedGroupId {
			_, err := client.NewGroupDelete().GroupID(group.ID).Do(context.Background())
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func cleanupConnectors(groupId string) {
	connectors, err := client.NewGroupListConnectors().GroupID(groupId).Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, connector := range connectors.Data.Items {
		_, err := client.NewConnectorDelete().ConnectorID(connector.ID).Do(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}
}

func cleanupWebhooks() {
	webhooks, err := client.NewWebhookList().Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, webhook := range webhooks.Data.Items {
		_, err := client.NewWebhookDelete().WebhookId(webhook.Id).Do(context.Background())
		if err != nil {
			log.Fatal(err)			
		}
	}
}