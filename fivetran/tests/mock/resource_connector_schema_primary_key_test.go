package mock

import (
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const primaryKeyTfConfig = `
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "connector_id"
	schema_change_handling = "BLOCK_ALL"
	schemas = {
		"public" = {
			enabled = true
			tables = {
				"users" = {
					enabled = true
					columns = {
						"id" = {
							enabled = true
							is_primary_key = true
						}
						"name" = {
							enabled = true
						}
					}
				}
			}
		}
	}
}
`

func primaryKeySchemaResponse(idIsPrimaryKey string) string {
	return `
{
	"enable_new_by_default": false,
	"schemas": {
		"public": {
			"name_in_destination": "public",
			"enabled": true,
			"tables": {
				"users": {
					"name_in_destination": "users",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					},
					"columns": {
						"id": {
							"name_in_destination": "id",
							"enabled": true,
							"is_primary_key": ` + idIsPrimaryKey + `,
							"enabled_patch_settings": {
								"allowed": true
							}
						},
						"name": {
							"name_in_destination": "name",
							"enabled": true,
							"is_primary_key": false,
							"enabled_patch_settings": {
								"allowed": true
							}
						}
					}
				}
			}
		}
	},
	"schema_change_handling": "BLOCK_ALL"
}
`
}

func setupPrimaryKeyMocks(t *testing.T, succeededAt interface{}, initialIsPrimaryKey string, patchBodies *[]map[string]interface{}) {
	var schemaData map[string]interface{}

	mockClient.Reset()

	mockClient.When(http.MethodGet, "/v1/connections/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", map[string]interface{}{
				"id":           "connector_id",
				"service":      "s3",
				"schema":       "public",
				"succeeded_at": succeededAt,
			}), nil
		},
	)

	mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			if schemaData == nil {
				schemaData = createMapFromJsonString(t, primaryKeySchemaResponse(initialIsPrimaryKey))
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
		},
	)

	mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			*patchBodies = append(*patchBodies, requestBodyToJson(t, req))
			schemaData = createMapFromJsonString(t, primaryKeySchemaResponse("true"))
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
		},
	)
}

func assertPatchSetsPrimaryKey(t *testing.T, patchBodies []map[string]interface{}) {
	t.Helper()
	assertNotEmpty(t, patchBodies)
	if len(patchBodies) == 0 {
		return
	}
	body := patchBodies[0]
	schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})
	public := assertKeyExists(t, schemas, "public").(map[string]interface{})
	tables := assertKeyExists(t, public, "tables").(map[string]interface{})
	users := assertKeyExists(t, tables, "users").(map[string]interface{})
	columns := assertKeyExists(t, users, "columns").(map[string]interface{})
	id := assertKeyExists(t, columns, "id").(map[string]interface{})
	assertKeyExistsAndHasValue(t, id, "is_primary_key", true)
}

// is_primary_key is sent to the API and read back into state for a not yet synced file connector
func TestResourceSchemaPrimaryKeyPreSync(t *testing.T) {
	var patchBodies []map[string]interface{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				patchBodies = nil
				setupPrimaryKeyMocks(t, nil, "false", &patchBodies)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: primaryKeyTfConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertPatchSetsPrimaryKey(t, patchBodies)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema",
							"schemas.public.tables.users.columns.id.is_primary_key", "true"),
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema",
							"schemas.public.tables.users.columns.name.is_primary_key", "false"),
					),
				},
			},
		},
	)
}

// the post-sync check is advisory: it warns (see changedPrimaryKeys unit tests) but doesn't block apply
func TestResourceSchemaPrimaryKeyPostSync(t *testing.T) {
	var patchBodies []map[string]interface{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				patchBodies = nil
				setupPrimaryKeyMocks(t, time.Now().Add(-24*time.Hour).Format(time.RFC3339), "false", &patchBodies)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: primaryKeyTfConfig,
					Check: resource.ComposeAggregateTestCheckFunc(
						func(s *terraform.State) error {
							assertPatchSetsPrimaryKey(t, patchBodies)
							return nil
						},
						resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema",
							"schemas.public.tables.users.columns.id.is_primary_key", "true"),
					),
				},
			},
		},
	)
}
