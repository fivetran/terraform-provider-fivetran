package resources_test

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	tfmock "github.com/fivetran/terraform-provider-fivetran/fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	groupUserGetHandler    *mock.Handler
	groupPostUserHandler   *mock.Handler
	groupDeleteUserHandler *mock.Handler
	groupGetUsersHandler   *mock.Handler
	groupUsersData         []interface{}
)

func setupMockClientGroupUsersResource(t *testing.T, initialUsers []interface{}) {
	tfmock.MockClient().Reset()

	groupUsersData = make([]interface{}, 0)

	if len(initialUsers) > 0 {
		groupUsersData = append(groupUsersData, initialUsers...)
	}

	var addedUserId = int(10)

	fetchUserIdFromURI := func(uri string) string {
		return strings.Split(uri, "/users/")[1]
	}

	groupUserGetHandler = tfmock.MockClient().When(http.MethodGet, "/v1/groups/group_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			var groupResponse = `{
				"id": "group_id",
				"name": "Group",
				"created_at": "2018-12-20T11:59:35.089589Z"
			}`
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK, "Success", tfmock.CreateMapFromJsonString(t, groupResponse)), nil
		},
	)

	groupGetUsersHandler = tfmock.MockClient().When(http.MethodGet, "/v1/groups/group_id/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := make(map[string]interface{})
			body["items"] = groupUsersData
			response := tfmock.FivetranSuccessResponse(t, req, http.StatusOK,
				"", body)
			return response, nil
		},
	)

	groupDeleteUserHandler = tfmock.MockClient().WhenWc(http.MethodDelete, "/v1/groups/group_id/users/*").ThenCall(
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
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK,
				fmt.Sprintf("User with id '%s' has been removed from the group", userId), nil), nil
		},
	)

	groupPostUserHandler = tfmock.MockClient().When(http.MethodPost, "/v1/groups/group_id/users").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := tfmock.RequestBodyToJson(t, req)
			tfmock.AssertKeyExists(t, body, "email")
			tfmock.AssertKeyExists(t, body, "role")

			// assign user id
			body["id"] = "user_" + strconv.Itoa(addedUserId)

			groupUsersData = append(groupUsersData, body)

			addedUserId++
			return tfmock.FivetranSuccessResponse(t, req, http.StatusOK,
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
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertNotEmpty(t, groupUsersData)
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
							tfmock.AssertEqual(t, groupGetUsersHandler.Interactions, 1)
							return nil
						},
						//resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
					),
				},
			},
		},
	)
}

func composeImportStateCheck(fs ...resource.ImportStateCheckFunc) resource.ImportStateCheckFunc {
	return func(s []*terraform.InstanceState) error {
		for i, f := range fs {
			if err := f(s); err != nil {
				return fmt.Errorf("check %d/%d error: %s", i+1, len(fs), err)
			}
		}

		return nil
	}
}

func testImportCheckResourceAttr(instanceId, attributeName, value string) resource.ImportStateCheckFunc {
	_, file, line, _ := runtime.Caller(1)

	return func(s []*terraform.InstanceState) error {
		for _, v := range s {
			if v.ID != instanceId {
				continue
			}

			if attrVal, ok := v.Attributes[attributeName]; ok {
				if attrVal != value {
					return fmt.Errorf("For %s with '%s' id, '%s' attribute value is expected: '%s', got: '%s'. At %s:%d", v.Ephemeral.Type, instanceId, attributeName, value, attrVal, file, line)
				}

				return nil
			} else {
				return fmt.Errorf("Attribute '%s' not found for %s with '%s' id. At %s:%d", attributeName, v.Ephemeral.Type, instanceId, file, line)
			}
		}

		return fmt.Errorf("Not found: %s with '%s' id. At %s:%d", s[0].Ephemeral.Type, instanceId, file, line)
	}
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
				tfmock.AssertEqual(t, groupGetUsersHandler.Interactions, 1)
				tfmock.AssertEqual(t, groupPostUserHandler.Interactions, 1)
				tfmock.AssertEqual(t, groupDeleteUserHandler.Interactions, 0)
				tfmock.AssertNotEmpty(t, groupUsersData)
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
				tfmock.AssertNotEmpty(t, groupUsersData)
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
				tfmock.AssertNotEmpty(t, groupUsersData)
				return nil
			},
			//resource.TestCheckResourceAttr("fivetran_group.testgroup", "name", "test_group_name"),
		),
	}
	step4 := resource.TestStep{
		ImportState:             true,
		ResourceName:            "fivetran_group_users.testgroup_users",
		ImportStateId:           "group_id",
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"last_updated"},
		ImportStateCheck: composeImportStateCheck(
			func(s []*terraform.InstanceState) error {
				tfmock.AssertEqual(t, len(groupUsersData), 2)
				return nil
			},
			testImportCheckResourceAttr("group_id", "user.#", "2"),
			testImportCheckResourceAttr("group_id", "user.0.id", "user_12"),
			testImportCheckResourceAttr("group_id", "user.0.email", "email1@user.domain"),
			testImportCheckResourceAttr("group_id", "user.0.role", "Read Only"),
			testImportCheckResourceAttr("group_id", "user.1.id", "user_10"),
			testImportCheckResourceAttr("group_id", "user.1.email", "email@user.domain"),
			testImportCheckResourceAttr("group_id", "user.1.role", "Destination Administrator"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientGroupUsersResource(t, nil)
			},
			ProtoV6ProviderFactories: tfmock.ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				tfmock.AssertEqual(t, groupDeleteUserHandler.Interactions, 3)
				tfmock.AssertEmpty(t, groupUsersData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
				step2,
				step3,
				step4,
			},
		},
	)
}
