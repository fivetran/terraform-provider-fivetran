package resources_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	groupGetHandler    *mock.Handler
	groupPostHandler   *mock.Handler
	groupPatchHandler  *mock.Handler
	groupDeleteHandler *mock.Handler
	groupData          map[string]interface{}
)

func onPostGroup(t *testing.T, req *http.Request) (*http.Response, error) {
	tfmock.AssertEmpty(t, groupData)

	body := tfmock.RequestBodyToJson(t, req)

	// Check the request
	tfmock.AssertEqual(t, len(body), 1)

	// Add response fields
	body["id"] = "group_id"
	body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")
	groupData = body

	response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated,
		"Group has been created", groupData)

	return response, nil
}

func onPatchGroup(t *testing.T, req *http.Request) (*http.Response, error) {
	tfmock.AssertNotEmpty(t, groupData)

	body := tfmock.RequestBodyToJson(t, req)

	// Check the request
	tfmock.AssertEqual(t, len(body), 1)

	// Update saved values
	for k, v := range body {
		groupData[k] = v
	}

	response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Group has been updated", groupData)

	return response, nil
}

func setupMockClientGroupResource(t *testing.T) {
	tfmock.MockClient().Reset()
	groupData = nil

	groupPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/groups").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPostGroup(t, req)
		},
	)

	groupGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			tfmock.AssertNotEmpty(t, groupData)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", groupData)
			return response, nil
		},
	)

	groupPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPatchGroup(t, req)
		},
	)

	groupDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			tfmock.AssertNotEmpty(t, groupData)
			groupData = nil
			response := tfmock.FivetranSuccessResponse(t, req, 200,
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
				tfmock.AssertEqual(t, groupPostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, groupData)
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
				tfmock.AssertEqual(t, groupPostHandler.Interactions, 2)
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
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, groupDeleteHandler.Interactions, 2)
				tfmock.AssertEmpty(t, groupData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)
}
