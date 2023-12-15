package mock

import (
	"net/http"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	userPostHandler   *mock.Handler
	userPatchHandler  *mock.Handler
	userDeleteHandler *mock.Handler
	userData          map[string]interface{}
)

func onPostUsers(t *testing.T, req *http.Request) (*http.Response, error) {
	assertEmpty(t, userData)

	body := requestBodyToJson(t, req)

	// Check the request
	assertEqual(t, len(body), 6)
	assertEqual(t, body["email"], "john.fox@testmail.com")
	assertEqual(t, body["given_name"], "John")
	assertEqual(t, body["family_name"], "Fox")
	assertEqual(t, body["phone"], "+19876543210")
	assertEqual(t, body["picture"], "https://myPicturecom")
	assertEqual(t, body["role"], "Account Reviewer")

	// Add response fields
	body["id"] = "john_fox_id"
	body["verified"] = false
	body["invited"] = true
	body["logged_in_at"] = nil
	body["created_at"] = time.Now().Format("2006-01-02T15:04:05.000000Z")
	userData = body

	response := fivetranSuccessResponse(t, req, http.StatusCreated,
		"User has been invited to the account", body)

	return response, nil
}

func onPatchUser(t *testing.T, req *http.Request, updateIteration int) (*http.Response, error) {
	assertNotEmpty(t, userData)

	body := requestBodyToJson(t, req)

	if updateIteration == 0 {
		// Check the request
		assertEqual(t, len(body), 5)
		assertEqual(t, body["given_name"], "Jane")
		assertEqual(t, body["family_name"], "Connor")
		assertEqual(t, body["phone"], "+19876543219")
		assertEqual(t, body["picture"], "https://yourPicturecom")
		assertEqual(t, body["role"], "Account Administrator")

		// Update saved values
		for k, v := range body {
			userData[k] = v
		}

		response := fivetranSuccessResponse(t, req, http.StatusOK, "User has been updated", userData)
		return response, nil
	}

	if updateIteration == 1 {
		// Check the request
		assertEqual(t, len(body), 2)
		assertEqual(t, body["phone"], nil)
		assertEqual(t, body["picture"], nil)

		// Update saved values
		for k, v := range body {
			userData[k] = v
		}

		response := fivetranSuccessResponse(t, req, http.StatusOK, "User has been updated", userData)
		return response, nil
	}

	return nil, nil
}

func setupMockClientUserResource(t *testing.T) {
	mockClient.Reset()
	userData = nil
	updateCounter := 0

	userPostHandler = mockClient.When(http.MethodPost, "/v1/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return onPostUsers(t, req)
		},
	)

	mockClient.When(http.MethodGet, "/v1/users/john_fox_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			assertNotEmpty(t, userData)
			response := fivetranSuccessResponse(t, req, http.StatusOK, "", userData)
			return response, nil
		},
	)

	userPatchHandler = mockClient.When(http.MethodPatch, "/v1/users/john_fox_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			response, err := onPatchUser(t, req, updateCounter)
			updateCounter++
			return response, err
		},
	)

	userDeleteHandler = mockClient.When(http.MethodDelete, "/v1/users/john_fox_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			assertNotEmpty(t, userData)
			userData = nil
			response := fivetranSuccessResponse(t, req, 200,
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
				assertEqual(t, userPostHandler.Interactions, 1)
				assertNotEmpty(t, userData)
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
				assertEqual(t, userPatchHandler.Interactions, 1)
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
				assertEqual(t, userPatchHandler.Interactions, 2)
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
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, userDeleteHandler.Interactions, 1)
				assertEmpty(t, userData)
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
