package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	groupUsersDataSourceMockGetHandler *mock.Handler
	groupUsersDataSourceMockData       map[string]interface{}
)

const (
	groupUsersMappingResponse = `
	{
        "items": [
            {
                "id": "user_id",
                "email": "john@mycompany.com",
                "given_name": "John",
                "family_name": "White",
                "verified": true,
                "invited": false,
                "picture": null,
                "phone": null,
                "logged_in_at": "2019-01-03T08:44:45.369Z",
                "created_at": "2018-01-15T11:00:27.329220Z",
                "active": true
            }
        ],
        "next_cursor": null
    }
	`
)

func setupMockClientGroupUsersDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	groupUsersDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/groups/group_id/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			groupUsersDataSourceMockData = createMapFromJsonString(t, groupUsersMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", groupUsersDataSourceMockData), nil
		},
	)
}

func TestDataSourceGroupUsersMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_group_users" "test_users" {
			provider = fivetran-provider
			id = "group_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, groupUsersDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, groupUsersDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.id", "user_id"),
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.email", "john@mycompany.com"),
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.given_name", "John"),
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.family_name", "White"),
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.verified", "true"),
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.invited", "false"),
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.picture", ""),
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.phone", ""),
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.logged_in_at", "2019-01-03 08:44:45.369 +0000 UTC"),
			resource.TestCheckResourceAttr("data.fivetran_group_users.test_users", "users.0.created_at", "2018-01-15 11:00:27.32922 +0000 UTC"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientGroupUsersDataSourceConfigMapping(t)
			},
			Providers: testProviders,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
