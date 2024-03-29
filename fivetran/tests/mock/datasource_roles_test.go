package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	rolesDataSourceMockGetHandler *mock.Handler
	rolesDataSourceMockData       map[string]interface{}
)

const (
	rolesMappingResponse = `
	{
    	"items": [
      	{
        	"name": "Account Administrator",
        	"description": "text_description",
        	"is_custom": false,
        	"scope": ["ACCOUNT"]
      	}],
    	"next_cursor": null
	}`
)

func setupMockClientRolesDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	rolesDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/roles").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			rolesDataSourceMockData = createMapFromJsonString(t, rolesMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", rolesDataSourceMockData), nil
		},
	)
}

func TestDataSourceRolesMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_roles" "test_roles" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, rolesDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, rolesDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_roles.test_roles", "roles.0.name", "Account Administrator"),
			resource.TestCheckResourceAttr("data.fivetran_roles.test_roles", "roles.0.description", "text_description"),
			resource.TestCheckResourceAttr("data.fivetran_roles.test_roles", "roles.0.is_custom", "false"),
			resource.TestCheckResourceAttr("data.fivetran_roles.test_roles", "roles.0.scope.0", "ACCOUNT"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientRolesDataSourceConfigMapping(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
