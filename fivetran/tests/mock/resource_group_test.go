package mock

import (
	"net/http"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	groupPostHandler   *mock.Handler
	groupPatchHandler  *mock.Handler
	groupDeleteHandler *mock.Handler
	groupData          map[string]interface{}
)

func onPostGroup(t *testing.T, req *http.Request) (*http.Response, error) {
	assertEmpty(t, groupData)

	body := requestBodyToJson(t, req)

	// Check the request
	assertEqual(t, len(body), 1)

	// Add response fields
	body["id"] = "group_id"
	body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")
	groupData = body

	response := fivetranSuccessResponse(t, req, http.StatusCreated,
		"Group has been created", groupData)

	return response, nil
}

func onPatchGroup(t *testing.T, req *http.Request) (*http.Response, error) {
	assertNotEmpty(t, groupData)

	body := requestBodyToJson(t, req)

	// Check the request
	assertEqual(t, len(body), 1)

	// Update saved values
	for k, v := range body {
		groupData[k] = v
	}

	response := fivetranSuccessResponse(t, req, http.StatusOK, "Group has been updated", groupData)

	return response, nil
}

func setupMockClientGroupResource(t *testing.T) {
	mockClient.Reset()
	groupData = nil

	groupPostHandler = mockClient.When(http.MethodPost, "/v1/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPostGroup(t, req)
		},
	)

	mockClient.When(http.MethodGet, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			assertNotEmpty(t, groupData)
			response := fivetranSuccessResponse(t, req, http.StatusOK, "", groupData)
			return response, nil
		},
	)

	groupPatchHandler = mockClient.When(http.MethodPatch, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPatchGroup(t, req)
		},
	)

	groupDeleteHandler = mockClient.When(http.MethodDelete, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			assertNotEmpty(t, groupData)
			groupData = nil
			response := fivetranSuccessResponse(t, req, 200,
				"Group with id 'group_id' has been deleted", nil)
			return response, nil
		},
	)
}

func TestResourceGroupMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_group" "testgroup" {
				provider = fivetran-provider
				name = "test_group_name"
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, groupPostHandler.Interactions, 1)
				assertNotEmpty(t, groupData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
		),
	}

	step2 := resource.TestStep{
		Config: `
			resource "fivetran_group" "testgroup" {
				provider = fivetran-provider
				name = "new_test_group_name"
			}
		`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, groupPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "new_test_group_name"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientGroupResource(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, groupDeleteHandler.Interactions, 1)
				assertEmpty(t, groupData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}
