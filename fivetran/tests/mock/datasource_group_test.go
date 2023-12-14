package mock

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	groupDataSourceMockGetHandler *mock.Handler
	groupDataSourceMockData       map[string]interface{}
)

const (
	groupMappingResponse = `
	{
        "id": "group_id",
        "name": "group_name",
		"created_at": "2018-12-20T11:59:35.089589Z"
    }
	`
)

func setupMockClientGroupDataSourceConfigMapping(t *testing.T) {
	mockClient.Reset()

	groupDataSourceMockGetHandler = mockClient.When(http.MethodGet, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			groupDataSourceMockData = createMapFromJsonString(t, groupMappingResponse)
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", groupDataSourceMockData), nil
		},
	)
}

func TestDataSourceGroupMappingMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
		data "fivetran_group" "test_group" {
			provider = fivetran-provider
			id = "group_id"
		}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, groupDataSourceMockGetHandler.Interactions, 2)
				assertNotEmpty(t, groupDataSourceMockData)
				return nil
			},
			resource.TestCheckResourceAttr("data.fivetran_group.test_group", "name", "group_name"),
			resource.TestCheckResourceAttr("data.fivetran_group.test_group", "created_at", "2018-12-20 11:59:35.089589 +0000 UTC"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientGroupDataSourceConfigMapping(t)
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
