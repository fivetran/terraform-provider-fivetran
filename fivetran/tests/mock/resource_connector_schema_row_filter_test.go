package mock

import (
	"encoding/json"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const rowFilterSchemaResponse = `
{
	"schema_change_handling": "ALLOW_ALL",
	"row_filtering_supported": true,
	"schemas": {
		"salesforce": {
			"name_in_destination": "salesforce",
			"enabled": true,
			"tables": {
				"Account": {
					"name_in_destination": "account",
					"enabled": true,
					"enabled_patch_settings": {"allowed": true}
				},
				"Contact": {
					"name_in_destination": "contact",
					"enabled": true,
					"enabled_patch_settings": {"allowed": true}
				}
			}
		}
	}
}
`

const rowFilterUpstreamFilter = `
{
	"name": "existing_filter",
	"description": "Filter configured outside of Terraform",
	"column_clauses": [{"column": "RecordTypeId", "column_type": "STRING", "operator": "EQUALS", "values": ["012X"]}]
}
`

func rowFilterJsonConfig(rowFilter string) string {
	table := `"Account": {"enabled": true}`
	if rowFilter != "" {
		table = `"Account": {"enabled": true, "row_filter": ` + rowFilter + `}`
	}
	return `
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "connector_id"
	schema_change_handling = "ALLOW_ALL"
	schemas_json = <<EOT
{"salesforce": {"enabled": true, "tables": {` + table + `}}}
EOT
}
`
}

func rowFilterTypedConfig(rowFilter string) string {
	return `
resource "fivetran_connector_schema_config" "test_schema" {
	provider = fivetran-provider
	connector_id = "connector_id"
	schema_change_handling = "ALLOW_ALL"
	schemas = {
		"salesforce" = {
			enabled = true
			tables = {
				"Account" = {
					enabled = true
					` + rowFilter + `
				}
			}
		}
	}
}
`
}

type rowFilterMock struct {
	schemaData      map[string]interface{}
	patchBodies     []map[string]interface{}
	ignoreRowFilter bool // mimics a connection without row filtering available
}

func (m *rowFilterMock) account() map[string]interface{} {
	schemas := m.schemaData["schemas"].(map[string]interface{})
	tables := schemas["salesforce"].(map[string]interface{})["tables"].(map[string]interface{})
	return tables["Account"].(map[string]interface{})
}

func (m *rowFilterMock) setAccountRowFilter(t *testing.T, rowFilter string) {
	if rowFilter == "" {
		delete(m.account(), "row_filter")
		return
	}
	m.account()["row_filter"] = createMapFromJsonString(t, rowFilter)
}

// normalizeRowFilter mimics the API: the default AND operator is omitted and missing column types are resolved
func normalizeRowFilter(filter map[string]interface{}) map[string]interface{} {
	if filter["operator"] == "AND" {
		delete(filter, "operator")
	}
	if clauses, ok := filter["column_clauses"].([]interface{}); ok {
		for _, c := range clauses {
			clause := c.(map[string]interface{})
			if _, ok := clause["column_type"]; !ok {
				clause["column_type"] = "STRING"
			}
		}
	}
	return filter
}

func (m *rowFilterMock) lastPatchedAccount(t *testing.T) map[string]interface{} {
	t.Helper()
	assertNotEmpty(t, m.patchBodies)
	if len(m.patchBodies) == 0 {
		return nil
	}
	body := m.patchBodies[len(m.patchBodies)-1]
	schemas := assertKeyExists(t, body, "schemas").(map[string]interface{})
	salesforce := assertKeyExists(t, schemas, "salesforce").(map[string]interface{})
	tables := assertKeyExists(t, salesforce, "tables").(map[string]interface{})
	return assertKeyExists(t, tables, "Account").(map[string]interface{})
}

func setupRowFilterMocks(t *testing.T, m *rowFilterMock, initialRowFilter string, patchError bool) {
	mockClient.Reset()
	m.schemaData = createMapFromJsonString(t, rowFilterSchemaResponse)
	m.patchBodies = nil
	m.setAccountRowFilter(t, initialRowFilter)

	mockClient.When(http.MethodGet, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", m.schemaData), nil
		},
	)

	mockClient.When(http.MethodPatch, "/v1/connections/connector_id/schemas").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			body := requestBodyToJson(t, req)
			m.patchBodies = append(m.patchBodies, body)
			if patchError {
				return fivetranResponse(t, req, "InvalidInput", http.StatusBadRequest,
					"Row filter column `RecordTypeId` not found", nil), nil
			}
			schemas, _ := body["schemas"].(map[string]interface{})
			for _, s := range schemas {
				tables, _ := s.(map[string]interface{})["tables"].(map[string]interface{})
				for tName, tValue := range tables {
					table := tValue.(map[string]interface{})
					if tName != "Account" {
						continue
					}
					if rowFilter, ok := table["row_filter"]; ok && !m.ignoreRowFilter {
						if rowFilter == nil {
							delete(m.account(), "row_filter")
						} else {
							// copy, the recorded request body has to stay as sent
							encoded, _ := json.Marshal(rowFilter)
							m.account()["row_filter"] = normalizeRowFilter(createMapFromJsonString(t, string(encoded)))
						}
					}
				}
			}
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", m.schemaData), nil
		},
	)
}

func checkRowFilterMock(f func() error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return f()
	}
}

// row_filter in schemas_json is sent to the API, read back without a diff and deleted by an explicit null
func TestResourceSchemaRowFilterJsonMock(t *testing.T) {
	m := &rowFilterMock{}
	filter := `{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId",
		"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]}]}`
	updatedFilter := `{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId", "operator": "OR",
		"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A", "012B"]}, {"column": "Name", "operator": "IS_NULL"}]}`

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupRowFilterMocks(t, m, "", false)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: rowFilterJsonConfig(filter),
					Check: resource.ComposeAggregateTestCheckFunc(
						checkRowFilterMock(func() error {
							account := m.lastPatchedAccount(t)
							rowFilter := assertKeyExists(t, account, "row_filter").(map[string]interface{})
							assertEqual(t, rowFilter["name"], "account_record_type_filter")
							assertKeyExists(t, m.account(), "row_filter")
							return nil
						}),
						resource.TestMatchResourceAttr("fivetran_connector_schema_config.test_schema", "schemas_json",
							regexp.MustCompile(`"row_filter":\s*\{\s*"name":\s*"account_record_type_filter"`)),
					),
				},
				{
					Config: rowFilterJsonConfig(updatedFilter),
					Check: checkRowFilterMock(func() error {
						rowFilter := assertKeyExists(t, m.lastPatchedAccount(t), "row_filter").(map[string]interface{})
						assertEqual(t, rowFilter["operator"], "OR")
						assertEqual(t, len(rowFilter["column_clauses"].([]interface{})), 2)
						return nil
					}),
				},
				{
					Config: rowFilterJsonConfig("null"),
					Check: checkRowFilterMock(func() error {
						account := m.lastPatchedAccount(t)
						assertKeyExistsAndHasValue(t, account, "row_filter", nil)
						AssertKeyDoesNotExist(t, m.account(), "row_filter")
						return nil
					}),
				},
			},
		},
	)
}

// a table without row_filter keeps the existing upstream filter
func TestResourceSchemaRowFilterOmittedPreservedMock(t *testing.T) {
	m := &rowFilterMock{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupRowFilterMocks(t, m, rowFilterUpstreamFilter, false)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: rowFilterJsonConfig(""),
					Check: resource.ComposeAggregateTestCheckFunc(
						checkRowFilterMock(func() error {
							for _, body := range m.patchBodies {
								schemas, _ := body["schemas"].(map[string]interface{})
								salesforce, _ := schemas["salesforce"].(map[string]interface{})
								tables, _ := salesforce["tables"].(map[string]interface{})
								account, _ := tables["Account"].(map[string]interface{})
								AssertKeyDoesNotExist(t, account, "row_filter")
							}
							assertKeyExists(t, m.account(), "row_filter")
							return nil
						}),
						resource.TestCheckResourceAttrWith("fivetran_connector_schema_config.test_schema", "schemas_json",
							func(value string) error {
								if regexp.MustCompile(`row_filter`).MatchString(value) {
									t.Errorf("expected unconfigured row_filter not to be saved in state, got %v", value)
								}
								return nil
							}),
					),
				},
			},
		},
	)
}

// row_filter in schemas is sent to the API, read back without a diff, and deleted when removed from the configuration
func TestResourceSchemaRowFilterTypedMock(t *testing.T) {
	m := &rowFilterMock{}
	filter := `row_filter = {
		name        = "account_record_type_filter"
		description = "Filter Account rows by RecordTypeId"
		operator    = "AND"
		column_clauses = [
			{
				column   = "RecordTypeId"
				operator = "IN"
				values   = ["012A"]
			},
			{
				column      = "Name"
				column_type = "STRING"
				operator    = "IS_NOT_NULL"
			}
		]
	}`
	updatedFilter := `row_filter = {
		name        = "account_record_type_filter"
		description = "Filter Account rows by RecordTypeId"
		column_clauses = [
			{
				column   = "RecordTypeId"
				operator = "EQUALS"
				values   = ["012B"]
			}
		]
	}`
	const resourceName = "fivetran_connector_schema_config.test_schema"

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupRowFilterMocks(t, m, "", false)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: rowFilterTypedConfig(filter),
					Check: resource.ComposeAggregateTestCheckFunc(
						checkRowFilterMock(func() error {
							rowFilter := assertKeyExists(t, m.lastPatchedAccount(t), "row_filter").(map[string]interface{})
							assertEqual(t, rowFilter["operator"], "AND")
							clauses := rowFilter["column_clauses"].([]interface{})
							assertEqual(t, len(clauses), 2)
							AssertKeyDoesNotExist(t, clauses[0].(map[string]interface{}), "column_type")
							AssertKeyDoesNotExist(t, clauses[1].(map[string]interface{}), "values")
							return nil
						}),
						resource.TestCheckResourceAttr(resourceName, "schemas.salesforce.tables.Account.row_filter.name", "account_record_type_filter"),
						resource.TestCheckResourceAttr(resourceName, "schemas.salesforce.tables.Account.row_filter.operator", "AND"),
						resource.TestCheckResourceAttr(resourceName, "schemas.salesforce.tables.Account.row_filter.column_clauses.0.values.0", "012A"),
						resource.TestCheckNoResourceAttr(resourceName, "schemas.salesforce.tables.Account.row_filter.column_clauses.0.column_type"),
						resource.TestCheckResourceAttr(resourceName, "schemas.salesforce.tables.Account.row_filter.column_clauses.1.column_type", "STRING"),
						resource.TestCheckNoResourceAttr(resourceName, "schemas.salesforce.tables.Contact.row_filter"),
					),
				},
				{
					Config: rowFilterTypedConfig(updatedFilter),
					Check: resource.ComposeAggregateTestCheckFunc(
						checkRowFilterMock(func() error {
							rowFilter := assertKeyExists(t, m.lastPatchedAccount(t), "row_filter").(map[string]interface{})
							assertEqual(t, len(rowFilter["column_clauses"].([]interface{})), 1)
							return nil
						}),
						resource.TestCheckResourceAttr(resourceName, "schemas.salesforce.tables.Account.row_filter.column_clauses.0.values.0", "012B"),
					),
				},
				{
					Config: rowFilterTypedConfig(""),
					Check: resource.ComposeAggregateTestCheckFunc(
						checkRowFilterMock(func() error {
							assertKeyExistsAndHasValue(t, m.lastPatchedAccount(t), "row_filter", nil)
							AssertKeyDoesNotExist(t, m.account(), "row_filter")
							return nil
						}),
						resource.TestCheckNoResourceAttr(resourceName, "schemas.salesforce.tables.Account.row_filter"),
					),
				},
			},
		},
	)
}

// a configured row filter changed outside of Terraform is detected as a difference
func TestResourceSchemaRowFilterDriftMock(t *testing.T) {
	m := &rowFilterMock{}
	filter := `{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId",
		"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]}]}`

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupRowFilterMocks(t, m, "", false)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: rowFilterJsonConfig(filter),
				},
				{
					PreConfig: func() {
						m.setAccountRowFilter(t, rowFilterUpstreamFilter)
					},
					Config:             rowFilterJsonConfig(filter),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
				{
					PreConfig: func() {
						m.setAccountRowFilter(t, "")
					},
					Config:             rowFilterJsonConfig(filter),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
				},
			},
		},
	)
}

// an API error while applying a row filter fails the apply
func TestResourceSchemaRowFilterApiErrorMock(t *testing.T) {
	m := &rowFilterMock{}
	filter := `{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId",
		"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]}]}`

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupRowFilterMocks(t, m, "", true)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      rowFilterJsonConfig(filter),
					ExpectError: regexp.MustCompile("Row filter column `RecordTypeId` not found"),
				},
			},
		},
	)
}

// an invalid row filter in schemas_json fails at plan time, before any request is sent
func TestResourceSchemaRowFilterJsonValidationMock(t *testing.T) {
	m := &rowFilterMock{}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupRowFilterMocks(t, m, "", false)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      rowFilterJsonConfig(`{"name": "account_record_type_filter", "column_clauses": []}`),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile("Invalid row filter for table `Account` of schema `salesforce`(?s:.*)`row_filter.description` is required"),
				},
				{
					Config:   rowFilterJsonConfig(""),
					PlanOnly: true,
					PreConfig: func() {
						assertEqual(t, len(m.patchBodies), 0)
					},
					ExpectNonEmptyPlan: true,
				},
			},
		},
	)
}

// switching from schemas to schemas_json in the same apply that adds a row filter sends the filter (issue #630)
func TestResourceSchemaRowFilterSwitchToJsonMock(t *testing.T) {
	m := &rowFilterMock{}
	filter := `{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId",
		"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]}]}`

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupRowFilterMocks(t, m, "", false)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: rowFilterTypedConfig(""),
				},
				{
					Config: rowFilterJsonConfig(filter),
					Check: checkRowFilterMock(func() error {
						rowFilter := assertKeyExists(t, m.lastPatchedAccount(t), "row_filter").(map[string]interface{})
						assertEqual(t, rowFilter["name"], "account_record_type_filter")
						assertKeyExists(t, m.account(), "row_filter")
						return nil
					}),
				},
			},
		},
	)
}

// a row filter accepted by the API but not saved fails the apply instead of being silently ignored
func TestResourceSchemaRowFilterNotAppliedMock(t *testing.T) {
	m := &rowFilterMock{}
	filter := `{"name": "account_record_type_filter", "description": "Filter Account rows by RecordTypeId",
		"column_clauses": [{"column": "RecordTypeId", "operator": "IN", "values": ["012A"]}]}`

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupRowFilterMocks(t, m, "", false)
				m.ignoreRowFilter = true
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      rowFilterJsonConfig(filter),
					ExpectError: regexp.MustCompile("Row filters were not applied(?s:.*)salesforce.Account(?s:.*)row_filtering_supported"),
				},
			},
		},
	)
}
