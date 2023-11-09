package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	userDataSourceMockGetHandler *mock.Handler
	userDataSourceMockData       map[string]interface{}
)

const (
	userMappingResponse = `
	{
        "id": "user_id",
        "email": "john@mycompany.com",
        "given_name": "John",
        "family_name": "White",
        "verified": true,
        "invited": false,
        "picture": "https://some.picture.url",
        "phone": "+123456789",
        "role": "Account Reviewer",
        "logged_in_at": "2019-01-03T08:44:45.369Z",
        "created_at": "2018-01-15T11:00:27.329220Z",
        "active": true
    }
	`
)

func setupMockClientUserDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	userDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/users/user_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userDataSourceMockData = createMapFromJsonString(t, userMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", userDataSourceMockData), nil
		},
	)
}

func TestDataSourceUserMappingMock(t *testing.T) {
	// NOTE: the config is totally inconsistent and contains all possible values for mapping test
	step1 := resource.TestStep{
		Config: `
		data "fivetran_user" "test_user" {
			provider = fivetran-provider
			id = "user_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, userDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, userDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_user.test_user", "email", "john@mycompany.com"),
			resource.TestCheckResourceAttr("data.fivetran_user.test_user", "given_name", "John"),
			resource.TestCheckResourceAttr("data.fivetran_user.test_user", "verified", "true"),
			resource.TestCheckResourceAttr("data.fivetran_user.test_user", "invited", "false"),
			resource.TestCheckResourceAttr("data.fivetran_user.test_user", "picture", "https://some.picture.url"),
			resource.TestCheckResourceAttr("data.fivetran_user.test_user", "phone", "+123456789"),
			resource.TestCheckResourceAttr("data.fivetran_user.test_user", "role", "Account Reviewer"),
			resource.TestCheckResourceAttr("data.fivetran_user.test_user", "logged_in_at", "2019-01-03 08:44:45.369 +0000 UTC"),
			resource.TestCheckResourceAttr("data.fivetran_user.test_user", "created_at", "2018-01-15 11:00:27.32922 +0000 UTC"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientUserDataSourceConfigMapping(t)
			},
			//Providers: testProviders,
			ProtoV5ProviderFactories: protoV5ProviderFactory,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
