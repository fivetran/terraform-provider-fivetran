package mock

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	groupGetHandler        *mock.Handler
	groupPostUserHandler   *mock.Handler
	groupDeleteUserHandler *mock.Handler
	groupGetUsersHandler   *mock.Handler
	groupUsersData         []interface{}
)

func setupMockClientGroupUsersResource(t *testing.T, initialUsers []interface{}) {
	mockClient.Reset()

	groupUsersData = make([]interface{}, 0)

	if len(initialUsers) > 0 {
		groupUsersData = append(groupUsersData, initialUsers...)
	}

	var addedUserId = int(10)

	fetchUserIdFromURI := func(uri string) string {
		return strings.Split(uri, "/users/")[1]
	}

	groupGetHandler = mockClient.When(http.MethodGet, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			var groupResponse = `{
				"id": "group_id",
				"name": "Group",
				"created_at": "2018-12-20T11:59:35.089589Z"
			}`
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", createMapFromJsonString(t, groupResponse)), nil
		},
	)

	groupGetUsersHandler = mockClient.When(http.MethodGet, "/v1/groups/group_id/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := make(map[string]interface{})
			body["items"] = groupUsersData
			response := fivetranSuccessResponse(t, req, http.StatusOK,
				"", body)
			return response, nil
		},
	)

	groupDeleteUserHandler = mockClient.WhenWc(http.MethodDelete, "/v1/groups/group_id/users/*").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			userId := fetchUserIdFromURI(req.URL.Path)
			newGroupUsersData := make([]interface{}, 0)
			for i, u := range groupUsersData {
				if u.(map[string]interface{})["id"].(string) != userId {
					newGroupUsersData = append(newGroupUsersData, u)
				} else {
					// once we have found iser with userId we can just append the rest users to result
					newGroupUsersData = append(newGroupUsersData, groupUsersData[i+1:]...)
					break
				}
			}
			groupUsersData = newGroupUsersData
			return fivetranSuccessResponse(t, req, http.StatusOK,
				fmt.Sprintf("User with id '%s' has been removed from the group", userId), nil), nil
		},
	)

	groupPostUserHandler = mockClient.When(http.MethodPost, "/v1/groups/group_id/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)
			assertKeyExists(t, body, "email")
			assertKeyExists(t, body, "role")

			// assign user id
			body["id"] = "user_" + strconv.Itoa(addedUserId)

			groupUsersData = append(groupUsersData, body)

			addedUserId++
			return fivetranSuccessResponse(t, req, http.StatusOK,
				"User has been added to the group", nil), nil
		},
	)
}

func TestResourceGroupUsersCleanupGroupOnCreate(t *testing.T) {
	initialUsers := make([]interface{}, 0)

	user := make(map[string]interface{})

	user["id"] = "initial_user"
	user["email"] = "initial_user@email"
	user["role"] = "Some Role"

	initialUsers = append(initialUsers, user)

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientGroupUsersResource(t, initialUsers)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				assertNotEmpty(t, groupUsersData)
				return nil
			},

			Steps: []resource.TestStep{
				{
					Config: `
						resource "fivetran_group_users" "testgroup_users" {
							provider = fivetran-provider
			
							group_id = "group_id"
						}`,

					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertEqual(t, groupGetUsersHandler.Interactions, 1)
							return nil
						},
						//resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
					),
				},
			},
		},
	)
}

func TestResourceGroupUsersMock(t *testing.T) {
	step1 := resource.TestStep{
		Config: `
			resource "fivetran_group_users" "testgroup_users" {
				provider = fivetran-provider

				group_id = "group_id"

				user {
					email = "email@user.domain"
					role = "Destination Administrator"
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, groupGetUsersHandler.Interactions, 2)
				assertEqual(t, groupPostUserHandler.Interactions, 1)
				assertEqual(t, groupDeleteUserHandler.Interactions, 0)
				assertNotEmpty(t, groupUsersData)
				return nil
			},
			//resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
		),
	}
	step2 := resource.TestStep{
		Config: `
			resource "fivetran_group_users" "testgroup_users" {
				provider = fivetran-provider

				group_id = "group_id"

				user {
					email = "email@user.domain"
					role = "Destination Administrator"
				}

				user {
					email = "email1@user.domain"
					role = "Destination Administrator"
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertNotEmpty(t, groupUsersData)
				return nil
			},
			//resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
		),
	}
	step3 := resource.TestStep{
		Config: `
			resource "fivetran_group_users" "testgroup_users" {
				provider = fivetran-provider

				group_id = "group_id"

				user {
					email = "email@user.domain"
					role = "Destination Administrator"
				}

				user {
					email = "email1@user.domain"
					role = "Read Only"
				}
			}`,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertNotEmpty(t, groupUsersData)
				return nil
			},
			//resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
		),
	}
	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientGroupUsersResource(t, nil)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, groupDeleteUserHandler.Interactions, 3)
				assertEmpty(t, groupUsersData)
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
