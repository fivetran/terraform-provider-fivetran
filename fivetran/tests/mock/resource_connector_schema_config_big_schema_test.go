package mock

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceSchemaRawWithExtraFieldsTestMock(t *testing.T) {
	// configs are identical from settings pov, only order of schemas chenged + values of unused fields.
	tfCofnig1 := `
	resource "fivetran_connector_schema_config" "test_schema" {
        provider = fivetran-provider
        connector_id = "connector_id"
        schema_change_handling = "BLOCK_ALL"
        schemas_json = <<EOT
        {
			"schema_0": {
				"enabled": true,
				"extra_field": "extra_value",
				"tables": {
					"table_0": {
						"extra_field": "extra_value",
						"enabled": true
					}

				}
			},
			"schema_2": {
				"enabled": true,
				"extra_field": "extra_value",
				"tables": {
					"table_0": {
						"extra_field": "extra_value",
						"enabled": true
					}

				}
			}
        }
        EOT
    }
	`

	tfCofnig2 := `
	resource "fivetran_connector_schema_config" "test_schema" {
        provider = fivetran-provider
        connector_id = "connector_id"
        schema_change_handling = "BLOCK_ALL"
        schemas_json = <<EOT
        {
			"schema_2": {
				"enabled": true,
				"extra_field": "extra_value",
				"tables": {
					"table_0": {
						"extra_field": "extra_value",
						"enabled": true
					}

				}
			},
			"schema_0": {
				"enabled": true,
				"extra_field": "extra_value1",
				"tables": {
					"table_0": {
						"extra_field": "extra_value1",
						"enabled": true
					}

				}
			}
        }
        EOT
    }
	`

	jsonResponse := `
	{
        "enable_new_by_default": false,
        "schemas": {
			"schema_0": {
				"name_in_destination": "schema_0",
				"enabled": true,
				"tables": {
					"table_0": {
							"name_in_destination": "table_0",
							"enabled": true,
							"enabled_patch_settings": {
									"allowed": true
							}
					}
				}
			},
			"schema_2": {
				"name_in_destination": "schema_2",
				"enabled": true,
				"tables": {
					"table_0": {
							"name_in_destination": "table_0",
							"enabled": true,
							"enabled_patch_settings": {
									"allowed": true
							}
					}
				}
			}
        },
        "schema_change_handling": "BLOCK_ALL"
	}
	`
	var schemaData map[string]interface{}

	var getHandler *mock.Handler

	step1 := resource.TestStep{
		Config: tfCofnig1,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, getHandler.Interactions, 1)
				assertNotEmpty(t, schemaData)
				return nil
			},
		),
	}

	step2 := resource.TestStep{
		Config: tfCofnig2,

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, getHandler.Interactions, 4)
				assertNotEmpty(t, schemaData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				schemaData = nil

				getHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							schemaData = createMapFromJsonString(t, jsonResponse)
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				step1,
				step2,
			},
		},
	)

}

func TestResourceSchemaBigSchemaRawTestMock(t *testing.T) {
	// check 10 schemas of 1000 tables
	schemasCount := 10
	tablesCount := 1000
	generateSchemasJson := func(schemaCount, tableCount int, response bool) string {
		tableTempate :=
			`
				"%v": {
					"name_in_destination": "%v",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
		`
		if !response {
			tableTempate =
				`
				"%v": {
					"enabled": true
				}
		`
		}
		schemaTemplate :=
			`
		"%v": {
			"name_in_destination": "%v",
			"enabled": true,
			"tables": {
%v
			}
		}
		`
		if !response {
			schemaTemplate =
				`
		"%v": {
			"enabled": true,
			"tables": {
%v
			}
		}
		`
		}
		schemas := ""
		for si := 0; si < schemaCount; si++ {
			if si > 0 {
				schemas = schemas + ","
			}
			tables := ""
			for ti := 0; ti < tableCount; ti++ {
				if ti > 0 {
					tables = tables + ","
				}
				tName := fmt.Sprintf("table_%v", ti)
				if response {
					tables = tables + fmt.Sprintf(tableTempate, tName, tName)
				} else {
					tables = tables + fmt.Sprintf(tableTempate, tName)
				}
			}
			sName := fmt.Sprintf("schema_%v", si)
			if response {
				schemas = schemas + fmt.Sprintf(schemaTemplate, sName, sName, tables)
			} else {
				schemas = schemas + fmt.Sprintf(schemaTemplate, sName, tables)
			}
		}
		return schemas
	}

	generateJsonResponse := func(schemaCount, tableCount int, sch string) string {
		schemas := generateSchemasJson(schemaCount, tableCount, true)
		result := fmt.Sprintf(
			`
{
	"enable_new_by_default": false,
	"schemas": {
%v
	},
	"schema_change_handling": "%v"
}
			`,
			schemas,
			sch,
		)
		fmt.Println("Response generation done")
		return result
	}

	generateTfConfig := func(schemaCount, tableCount int, sch string) string {
		start := time.Now()
		schemas := generateSchemasJson(schemaCount, tableCount, false)
		result := fmt.Sprintf(
			`
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "connector_id"
	schema_change_handling = "%v"
	schemas_json = <<EOT
	{
		%v
	}
	EOT
}
			`,
			sch,
			schemas,
		)
		fmt.Printf("Config generation done in %v\n", time.Since(start))
		return result
	}

	var schemaData map[string]interface{}

	var getHandler *mock.Handler

	step1 := resource.TestStep{
		Config: generateTfConfig(schemasCount, tablesCount, "BLOCK_ALL"),

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, getHandler.Interactions, 1)
				assertNotEmpty(t, schemaData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				schemaData = nil

				getHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							schemaData = createMapFromJsonString(t, generateJsonResponse(schemasCount, tablesCount, "BLOCK_ALL"))
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}

func TestResourceSchemaBigSchemaTestMock(t *testing.T) {

	schemasCount := 100
	tablesCount := 100

	generateJsonResponse := func(schemaCount, tableCount int, sch string) string {
		tableTempate :=
			`
				"%v": {
					"name_in_destination": "%v",
					"enabled": true,
					"enabled_patch_settings": {
						"allowed": true
					}
				}
		`
		schemaTemplate :=
			`
		"%v": {
			"name_in_destination": "%v",
			"enabled": true,
			"tables": {
%v
			}
		}
		`
		schemas := ""
		for si := 0; si < schemaCount; si++ {
			if si > 0 {
				schemas = schemas + ","
			}
			tables := ""
			for ti := 0; ti < tableCount; ti++ {
				if ti > 0 {
					tables = tables + ","
				}
				tName := fmt.Sprintf("table_%v", ti)
				tables = tables + fmt.Sprintf(tableTempate, tName, tName)
			}
			sName := fmt.Sprintf("schema_%v", si)
			schemas = schemas + fmt.Sprintf(schemaTemplate, sName, sName, tables)
		}
		result := fmt.Sprintf(
			`
{
	"enable_new_by_default": false,
	"schemas": {
%v
	},
	"schema_change_handling": "%v"
}
			`,
			schemas,
			sch,
		)
		fmt.Println("Response generation done")
		return result
	}

	generateTfConfig := func(schemaCount, tableCount int, sch string) string {
		tableTempate :=
			`
			"%v" = {
				enabled = true
			}
		`
		schemaTemplate :=
			`
	"%v" = {
		enabled = true
		tables = {
%v
		}
	}
		`
		schemas := ""
		for si := 0; si < schemaCount; si++ {
			tables := ""
			for ti := 0; ti < tableCount; ti++ {
				tName := fmt.Sprintf("table_%v", ti)
				tables = tables + fmt.Sprintf(tableTempate, tName)
			}
			sName := fmt.Sprintf("schema_%v", si)
			schemas = schemas + fmt.Sprintf(schemaTemplate, sName, tables)
		}
		result := fmt.Sprintf(
			`
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "connector_id"
	schema_change_handling = "%v"
	schemas = {
%v
	}
}
			`,
			sch,
			schemas,
		)
		fmt.Println("Config generation done")
		return result
	}

	var schemaData map[string]interface{}

	var getHandler *mock.Handler
	//var patchHandler *mock.Handler

	step1 := resource.TestStep{
		Config: generateTfConfig(schemasCount, tablesCount, "BLOCK_ALL"),

		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, getHandler.Interactions, 1)
				//assertEqual(t, patchHandler.Interactions, 1)
				assertNotEmpty(t, schemaData)
				return nil
			},
			//resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema_change_handling", "ALLOW_COLUMNS"),
			//resource.TestCheckResourceAttr("fivetran_connector_schema_config.test_schema", "schema.0.table.0.column.0.enabled", "false"),
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				mockClient.Reset()
				schemaData = nil

				getHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
					func(req *http.Request) (*http.Response, error) {
						if nil == schemaData {
							schemaData = createMapFromJsonString(t, generateJsonResponse(schemasCount, tablesCount, "BLOCK_ALL"))
						}
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success", schemaData), nil
					},
				)

			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it alsways exists within the connector
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
