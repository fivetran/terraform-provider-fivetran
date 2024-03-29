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
	userPostHandler   *mock.Handler
	userPatchHandler  *mock.Handler
	userDeleteHandler *mock.Handler
	userData          map[string]interface{}
)

func onPostUsers(t *testing.T, req *http.Request) (*http.Response, error) {
	tfmock.AssertEmpty(t, userData)

	body := tfmock.RequestBodyToJson(t, req)

	// Check the request
	tfmock.AssertEqual(t, len(body), 6)
	tfmock.AssertEqual(t, body["email"], "john.fox@testmail.com")
	tfmock.AssertEqual(t, body["given_name"], "John")
	tfmock.AssertEqual(t, body["family_name"], "Fox")
	tfmock.AssertEqual(t, body["phone"], "+19876543210")
	tfmock.AssertEqual(t, body["picture"], "https://myPicturecom")
	tfmock.AssertEqual(t, body["role"], "Account Reviewer")

	// Add response fields
	body["id"] = "john_fox_id"
	body["verified"] = false
	body["invited"] = true
	body["logged_in_at"] = nil
	body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")
	userData = body

	response := tfmock.FivetranSuccessResponse(t, req, http.StatusCreated,
		"User has been invited to the account", body)

	return response, nil
}

func onPatchUser(t *testing.T, req *http.Request, updateIteration int) (*http.Response, error) {
	tfmock.AssertNotEmpty(t, userData)

	body := tfmock.RequestBodyToJson(t, req)

	if updateIteration == 0 {
		// Check the request
		tfmock.AssertEqual(t, len(body), 5)
		tfmock.AssertEqual(t, body["given_name"], "Jane")
		tfmock.AssertEqual(t, body["family_name"], "Connor")
		tfmock.AssertEqual(t, body["phone"], "+19876543219")
		tfmock.AssertEqual(t, body["picture"], "https://yourPicturecom")
		tfmock.AssertEqual(t, body["role"], "Account Administrator")

		// Update saved values
		for k, v := range body {
			userData[k] = v
		}

		response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "User has been updated", userData)
		return response, nil
	}

	if updateIteration == 1 {
		// Check the request
		tfmock.AssertEqual(t, len(body), 2)
		tfmock.AssertEqual(t, body["phone"], nil)
		tfmock.AssertEqual(t, body["picture"], nil)

		// Update saved values
		for k, v := range body {
			userData[k] = v
		}

		response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "User has been updated", userData)
		return response, nil
	}

	return nil, nil
}

func setupMockClientUserResource(t *testing.T) {
	tfmock.MockClient().Reset()
	userData = nil
	updateCounter := 0

	userPostHandler = tfmock.MockClient().When(http.MethodPost, "/v1/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPostUsers(t, req)
		},
	)

	tfmock.MockClient().When(http.MethodGet, "/v1/users/john_fox_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			tfmock.AssertNotEmpty(t, userData)
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "", userData)
			return response, nil
		},
	)

	userPatchHandler = tfmock.MockClient().When(http.MethodPatch, "/v1/users/john_fox_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			response, err := onPatchUser(t, req, updateCounter)
			updateCounter++
			return response, err
		},
	)

	userDeleteHandler = tfmock.MockClient().When(http.MethodDelete, "/v1/users/john_fox_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			tfmock.AssertNotEmpty(t, userData)
			userData = nil
			response := tfmock.FivetranSuccessResponse(t, req, 200,
				"User with id 'john_fox_id' has been deleted", nil)
			return response, nil
		},
	)

}

func TestResourceUserMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_user" "userjohn" {
				provider = fivetran-provider
				email = "john.fox@testmail.com"
				family_name = "Fox"
				given_name = "John"
				role = "Account Reviewer"
				phone = "+19876543210"
				picture = "https://myPicturecom"
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, userPostHandler.Interactions, 1)
				tfmock.AssertNotEmpty(t, userData)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "email", "john.fox@testmail.com"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "role", "Account Reviewer"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "family_name", "Fox"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "given_name", "John"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "phone", "+19876543210"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "picture", "https://myPicturecom"),
		),
	}

	step2 := resource.TestStep{
		Config: `
			resource "fivetran_user" "userjohn" {
				provider = fivetran-provider
				role = "Account Administrator"
				email = "john.fox@testmail.com"
				family_name = "Connor"
				given_name = "Jane"
				phone = "+19876543219"
				picture = "https://yourPicturecom"
			}
		`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, userPatchHandler.Interactions, 1)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "email", "john.fox@testmail.com"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "role", "Account Administrator"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "family_name", "Connor"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "given_name", "Jane"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "phone", "+19876543219"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "picture", "https://yourPicturecom"),
		),
	}

	step3 := resource.TestStep{
		Config: `
			resource "fivetran_user" "userjohn" {
				provider = fivetran-provider
				role = "Account Administrator"
				email = "john.fox@testmail.com"
				family_name = "Connor"
				given_name = "Jane"
			}
		`,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				tfmock.AssertEqual(t, userPatchHandler.Interactions, 2)
				return nil
			},
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "email", "john.fox@testmail.com"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "role", "Account Administrator"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "family_name", "Connor"),
			resource.TestCheckResourceAttr("fivetran_user.userjohn", "given_name", "Jane"),
			resource.TestCheckNoResourceAttr("fivetran_user.userjohn", "phone"),
			resource.TestCheckNoResourceAttr("fivetran_user.userjohn", "picture"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientUserResource(t)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, userDeleteHandler.Interactions, 1)
				tfmock.AssertEmpty(t, userData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
			},
		},
	)
}
