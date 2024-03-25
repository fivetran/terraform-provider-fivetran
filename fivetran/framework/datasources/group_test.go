package datasources_test

import (
	"net/http"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	tfmock.MockClient().Reset()

	groupDataSourceMockGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			groupDataSourceMockData = tfmock.CreateMapFromJsonString(t, groupMappingResponse)
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", groupDataSourceMockData), nil
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
				tfmock.AssertEqual(t, groupDataSourceMockGetHandler.Interactions, 2)
				tfmock.AssertNotEmpty(t, groupDataSourceMockData)
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
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				return nil
			},
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
