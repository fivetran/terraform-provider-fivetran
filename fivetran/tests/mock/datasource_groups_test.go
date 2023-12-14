package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	groupsDataSourceMockGetHandler *mock.Handler
	groupsDataSourceMockData       map[string]interface{}
)

const (
	groupsMappingResponse = `
	{
		"items":[
			{
				"id": "group_id",
				"name": "group_name",
				"created_at": "2018-12-20T11:59:35.089589Z"
			}
		],
		"next_cursor": null	
    }
	`
)

func setupMockClientGroupsDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	groupsDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			groupsDataSourceMockData = createMapFromJsonString(t, groupsMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", groupsDataSourceMockData), nil
		},
	)
}

func TestDataSourceGroupsMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_groups" "test_groups" {
			provider = fivetran-provider
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, groupsDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, groupsDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_groups.test_groups", "groups.0.id", "group_id"),
			resource.TestCheckResourceAttr("data.fivetran_groups.test_groups", "groups.0.name", "group_name"),
			resource.TestCheckResourceAttr("data.fivetran_groups.test_groups", "groups.0.created_at", "2018-12-20 11:59:35.089589 +0000 UTC"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientGroupsDataSourceConfigMapping(t)
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
