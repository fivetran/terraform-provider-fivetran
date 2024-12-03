package mock

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type columnTestData struct {
	enabled              bool
	hashed               *bool
	enabledPatchSettings bool
	isPrimaryKey         bool
}

type tableTestData struct {
	enabled              bool
	syncMode             *string
	enabledPatchSettings bool
	columns              map[string]*columnTestData
}

type schemaTestData struct {
	enabled bool
	tables  map[string]*tableTestData
}

type schemaConfigTestData struct {
	schemaChangeHandling string
	schemas              map[string]*schemaTestData
}

type columnsConfigTestData struct {
	columns map[string]*columnTestData
}

func newColumnConfigTestData() columnsConfigTestData {
	return columnsConfigTestData{}
}

func (cctd columnsConfigTestData) newColumn(name string, enabled bool, hashed *bool, isPrimaryKey bool) columnsConfigTestData {
	if cctd.columns == nil {
		cctd.columns = map[string]*columnTestData{}
	}
	cctd.columns[name] = &columnTestData{
		enabled:              enabled,
		hashed:               hashed,
		enabledPatchSettings: true,
		isPrimaryKey:         isPrimaryKey,
	}
	return cctd
}

func (cctd columnsConfigTestData) jsonResponse() string {
	columns := ""
	for cName, c := range cctd.columns {
		if len(columns) > 0 {
			columns = columns + ",\n"
		}
		columns = columns + fmt.Sprintf(
			`    "%v": %v`, cName, c.jsonConfig())
	}
	columns = fmt.Sprintf(`
{
  "columns": {
%v  
  }
}`, columns)
	return columns
}

func (sctd *schemaConfigTestData) tfConfig() string {
	schemas := ""
	for sName, s := range sctd.schemas {
		schemas = schemas + fmt.Sprintf(`    "%v" = %v
`, sName, s.tfConfig())
	}
	if len(schemas) > 0 {
		schemas = fmt.Sprintf(`
  schemas = {
%v  }`, schemas)
	}
	return fmt.Sprintf(`
resource "fivetran_connector_schema_config" "test_schema" {
  provider = fivetran-provider
  connector_id = "connector_id"
  schema_change_handling = "%v"
  %v
}`, sctd.schemaChangeHandling, schemas)
}

func (sctd *schemaConfigTestData) tfConfigWithJsonSchemas() string {
	schemas := ""
	for sName, s := range sctd.schemas {
		if len(schemas) > 0 {
			schemas = schemas + ",\n"
		}
		schemas = schemas + fmt.Sprintf(`    "%v": %v`, sName, s.jsonConfig())
	}
	if len(schemas) > 0 {
		schemas = fmt.Sprintf(`
  schemas_json = <<EOT
  {
%v  
  }
  EOT`, schemas)
	}
	return fmt.Sprintf(`
resource "fivetran_connector_schema_config" "test_schema" {
  provider = fivetran-provider
  connector_id = "connector_id"
  schema_change_handling = "%v"
  %v
}`, sctd.schemaChangeHandling, schemas)
}

func (sctd *schemaConfigTestData) jsonResponse() string {
	schemas := ""
	for sName, s := range sctd.schemas {
		if len(schemas) > 0 {
			schemas = schemas + ",\n"
		}
		schemas = schemas + fmt.Sprintf(`    "%v": %v`, sName, s.jsonConfig())
	}
	if len(schemas) > 0 {
		schemas = fmt.Sprintf(`
  "schemas": {
%v  
  }`, schemas)
	}
	return fmt.Sprintf(`
{
  "schema_change_handling": "%v",
  %v
}`, sctd.schemaChangeHandling, schemas)
}

func (sctd *schemaConfigTestData) newSchema(name string, enabled bool) *schemaTestData {
	if sctd.schemas == nil {
		sctd.schemas = map[string]*schemaTestData{}
	}
	schema := &schemaTestData{
		enabled: enabled,
	}
	sctd.schemas[name] = schema
	return schema
}

func (std *schemaTestData) newTable(name string, enabled bool, syncMode *string) *tableTestData {
	if std.tables == nil {
		std.tables = map[string]*tableTestData{}
	}
	table := &tableTestData{
		enabled:              enabled,
		syncMode:             syncMode,
		enabledPatchSettings: true,
	}
	std.tables[name] = table
	return table
}

func (std *schemaTestData) newTableLocked(name string, enabled bool, syncMode *string) *tableTestData {
	if std.tables == nil {
		std.tables = map[string]*tableTestData{}
	}
	table := &tableTestData{
		enabled:              enabled,
		syncMode:             syncMode,
		enabledPatchSettings: false,
	}
	std.tables[name] = table
	return table
}

func (ttd *tableTestData) newColumn(name string, enabled bool, hashed *bool, isPrimaryKey bool) *tableTestData {
	if ttd.columns == nil {
		ttd.columns = map[string]*columnTestData{}
	}
	ttd.columns[name] = &columnTestData{
		enabled:              enabled,
		hashed:               hashed,
		enabledPatchSettings: true,
		isPrimaryKey:         isPrimaryKey,
	}
	return ttd
}

func (ttd *tableTestData) newColumnLocked(name string, enabled bool, hashed *bool, isPrimaryKey bool) *tableTestData {
	if ttd.columns == nil {
		ttd.columns = map[string]*columnTestData{}
	}
	ttd.columns[name] = &columnTestData{
		enabled:              enabled,
		hashed:               hashed,
		enabledPatchSettings: false,
		isPrimaryKey:         isPrimaryKey,
	}
	return ttd
}

func (ctd *columnTestData) tfConfig() string {
	template := `{
              enabled = %v %v
            }`
	hashed := ""
	if ctd.hashed != nil {
		hashed = fmt.Sprintf(`
              hashed = %v`, *ctd.hashed)
	}
	result := fmt.Sprintf(template, ctd.enabled, hashed)
	return result
}

func (ctd *columnTestData) jsonConfig() string {
	template := `{
		"enabled": %v %v %v
	  }`
	hashed := ""
	if ctd.hashed != nil {
		hashed = fmt.Sprintf(`,
              "hashed": %v`, *ctd.hashed)
	}
	patchSettings := fmt.Sprintf(`,
              "enabled_patch_settings": {
                "allowed": %v
              }
`, ctd.enabledPatchSettings)

	result := fmt.Sprintf(template, ctd.enabled, patchSettings, hashed)
	return result
}

func (ttd *tableTestData) tfConfig() string {
	tableTemplate := `{
          enabled = %v %v %v
        }`
	syncMode := ""
	if ttd.syncMode != nil && len(*ttd.syncMode) > 0 {
		syncMode = fmt.Sprintf(`
          sync_mode = "%v"`, *ttd.syncMode)
	}
	columns := ""
	for cName, c := range ttd.columns {
		columns = columns + fmt.Sprintf(`            "%v" = %v
`, cName, c.tfConfig())
	}
	if len(columns) > 0 {
		columns = fmt.Sprintf(`
          columns = {
%v          }`, columns)
	}
	result := fmt.Sprintf(tableTemplate, ttd.enabled, syncMode, columns)
	return result
}

func (ttd *tableTestData) jsonConfig() string {
	tableTemplate := `{
          "enabled": %v %v %v %v
        }`
	syncMode := ""
	if ttd.syncMode != nil && len(*ttd.syncMode) > 0 {
		syncMode = fmt.Sprintf(`,
          "sync_mode": "%v"`, *ttd.syncMode)
	}
	columns := ""
	for cName, c := range ttd.columns {
		if len(columns) > 0 {
			columns = columns + ",\n"
		}
		columns = columns + fmt.Sprintf(`            "%v": %v`, cName, c.jsonConfig())
	}
	if len(columns) > 0 {
		columns = fmt.Sprintf(`,
          "columns": {
%v
          }`, columns)
	}
	patchSettings := fmt.Sprintf(`,
          "enabled_patch_settings": {
            "allowed": %v
          }
`, ttd.enabledPatchSettings)

	result := fmt.Sprintf(tableTemplate, ttd.enabled, patchSettings, syncMode, columns)
	return result
}

func (std *schemaTestData) tfConfig() string {
	schemaTemplate := `{
      enabled = %v %v
    }`
	tables := ""
	for tName, t := range std.tables {
		tables = tables + fmt.Sprintf(`        "%v" = %v
`, tName, t.tfConfig())
	}
	if len(tables) > 0 {
		tables = fmt.Sprintf(`
      tables = {
%v      }`, tables)
	}
	result := fmt.Sprintf(schemaTemplate, std.enabled, tables)
	return result
}

func (std *schemaTestData) jsonConfig() string {
	schemaTemplate := `{
      "enabled": %v %v
    }`
	tables := ""
	for tName, t := range std.tables {
		if len(tables) > 0 {
			tables = tables + ",\n"
		}
		tables = tables + fmt.Sprintf(`        "%v": %v`, tName, t.jsonConfig())
	}
	if len(tables) > 0 {
		tables = fmt.Sprintf(`,
      "tables": {
%v
      }`, tables)
	}
	result := fmt.Sprintf(schemaTemplate, std.enabled, tables)
	return result
}

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func setupComplexTest(t *testing.T, initalUpstreamConfig schemaConfigTestData, tfConfigs, responseConfigs []schemaConfigTestData) []map[string]interface{} {
	var upstreamData map[string]interface{}
	bodies := []map[string]interface{}{}

	setupMockClient := func(t *testing.T) {
		mockClient.Reset()
		upstreamData = nil
		updateIteration := 0

		mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if nil == upstreamData {
					upstreamData = createMapFromJsonString(t, initalUpstreamConfig.jsonResponse())
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstreamData), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connectors/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				bodies = append(bodies, body)
				upstreamData = createMapFromJsonString(t, responseConfigs[updateIteration].jsonResponse())
				updateIteration++
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstreamData), nil
			},
		)
	}

	steps := []resource.TestStep{}
	for _, tfConfig := range tfConfigs {
		steps = append(steps, resource.TestStep{Config: tfConfig.tfConfig()})
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClient(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it alsways exists within the connector
				return nil
			},

			Steps: steps,
		},
	)

	return bodies
}

func setupComplexTestWithColumnsReload(
	t *testing.T,
	initalUpstreamConfig schemaConfigTestData,
	tfConfigs, responseConfigs []schemaConfigTestData,
	columnsResponseConfigs map[string](map[string]([]columnsConfigTestData)),
) []map[string]interface{} {
	var upstreamData map[string]interface{}
	bodies := []map[string]interface{}{}

	setupMockClient := func(t *testing.T) {
		mockClient.Reset()
		upstreamData = nil
		updateIteration := 0

		for schema, tables := range columnsResponseConfigs {
			for table, columnConfigs := range tables {
				mockClient.When(http.MethodGet, fmt.Sprintf("/v1/connectors/connector_id/schemas/%s/tables/%s/columns", schema, table)).ThenCall(
					func(req *http.Request) (*http.Response, error) {
						return fivetranSuccessResponse(t, req, http.StatusOK, "Success",
							createMapFromJsonString(t, columnConfigs[updateIteration].jsonResponse())), nil
					},
				)
			}
		}

		mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if nil == upstreamData {
					upstreamData = createMapFromJsonString(t, initalUpstreamConfig.jsonResponse())
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstreamData), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connectors/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body := requestBodyToJson(t, req)
				bodies = append(bodies, body)
				upstreamData = createMapFromJsonString(t, responseConfigs[updateIteration].jsonResponse())
				updateIteration++
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstreamData), nil
			},
		)
	}

	steps := []resource.TestStep{}
	for _, tfConfig := range tfConfigs {
		steps = append(steps, resource.TestStep{Config: tfConfig.tfConfig()})
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClient(t)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				// there is no possibility to destroy schema config - it alsways exists within the connector
				return nil
			},

			Steps: steps,
		},
	)

	return bodies
}

// Function performs test with create operation for resource (1 step) and returns request body
func setupOneStepTest(t *testing.T, upstreamConfig, tfConfig, responseConfig schemaConfigTestData) map[string]interface{} {
	var upstreamData map[string]interface{}
	var body map[string]interface{}

	setupMockClient := func(t *testing.T) {
		mockClient.Reset()
		upstreamData = nil

		mockClient.When(http.MethodGet, "/v1/connectors/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				if nil == upstreamData {
					upstreamData = createMapFromJsonString(t, upstreamConfig.jsonResponse())
				}
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstreamData), nil
			},
		)

		mockClient.When(http.MethodPatch, "/v1/connectors/connector_id/schemas").ThenCall(
			func(req *http.Request) (*http.Response, error) {
				body = requestBodyToJson(t, req)
				upstreamData = createMapFromJsonString(t, responseConfig.jsonResponse())
				return fivetranSuccessResponse(t, req, http.StatusOK, "Success", upstreamData), nil
			},
		)
	}
	step1 := resource.TestStep{
		Config: tfConfig.tfConfig(),
	}
	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClient(t)
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
	return body
}
