package mock

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/fivetran/go-fivetran/tests/mock"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var (
	connectorConflictingMockGetHandler  *mock.Handler
	connectorConflictingMockPostHandler *mock.Handler
	connectorConflictingMockDelete      *mock.Handler
	connectorConflictingMappingMockData map[string]interface{}
)

const (
	connectorConfigConflictingMappingTfConfig = `
resource "fivetran_connector" "test_connector" {
	provider = fivetran-provider

	group_id = "group_id"
	service = "%v"

	destination_schema {
		%v
	}

	trust_certificates = false
	trust_fingerprints = false
	run_setup_tests = false

	config {
		%v
	}
}
	`

	connectorConflictingMappingResponse = `
{
	"id": "connector_id",
	"group_id": "group_id",
	"service": "%v",
	"service_version": 1,
	"schema": "%v",
	"paused": true,
	"pause_after_trial": true,
	"connected_by": "user_id",
	"created_at": "2022-01-01T11:22:33.012345Z",
	"succeeded_at": null,
	"failed_at": null,
	"sync_frequency": 5,
	"schedule_type": "auto",
	"networking_method": "Directly",
	"status": {
		"setup_state": "incomplete",
		"sync_state": "paused",
		"update_state": "on_schedule",
		"is_historical_sync": true,
		"tasks": [{
			"code":"task_code",
			"message":"task_message"
		}],
		"warnings": [{
			"code":"warning_code",
			"message":"warning_message"
		}]
	},
	"setup_tests": [{
		"title": "Validate Login",
		"status": "FAILED",
		"message": "Invalid login credentials"
	}],
	"config": {
		%v
	}
}
	`
	debug = false
)

func getTfConfigForConflictingFields(service, destinationSchema, configTf string) string {
	return fmt.Sprintf(connectorConfigConflictingMappingTfConfig, service, destinationSchema, configTf)
}

func getJsonConfigForConflictingFields(service, schema, configJson string) string {
	return fmt.Sprintf(connectorConflictingMappingResponse, service, schema, configJson)
}

func setupMockClientConnectorResourceConfigConflictingFieldsMapping(t *testing.T, service, schema, configJson string) {
	mockClient.Reset()

	connectorConflictingMockGetHandler = mockClient.When(http.MethodGet, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorConflictingMappingMockData), nil
		},
	)

	connectorConflictingMockPostHandler = mockClient.When(http.MethodPost, "/v1/connectors").ThenCall(
		func(req *http.Request) (*http.Response, error) {

			response := getJsonConfigForConflictingFields(service, schema, configJson)

			if debug {
				fmt.Print(response)
			}

			connectorConflictingMappingMockData = createMapFromJsonString(t, response)
			return fivetranSuccessResponse(t, req, http.StatusCreated, "Success", connectorConflictingMappingMockData), nil
		},
	)

	connectorConflictingMockDelete = mockClient.When(http.MethodDelete, "/v1/connectors/connector_id").ThenCall(
		func(req *http.Request) (*http.Response, error) {
			connectorConflictingMappingMockData = nil
			return fivetranSuccessResponse(t, req, http.StatusOK, "Success", connectorConflictingMappingMockData), nil
		},
	)
}

func getTfConfigForField(fieldName, service string) string {
	if f, ok := common.GetConfigFieldsMap()[fieldName]; ok {
		return getTfConfigForFieldImpl(fieldName, service, f)
	}
	return ""
}

func getTfConfigForFieldImpl(fieldName, service string, field common.ConfigField) string {
	switch field.FieldValueType {
	case common.String:
		if field.ItemType != nil {
			if t, ok := field.ItemType[service]; ok {
				switch t {
				case common.Integer:
					return fmt.Sprintf(`%v = %v`, fieldName, "123")
				case common.Float:
					return fmt.Sprintf(`%v = %v`, fieldName, "123.4")
				}
			}
		}
		return fmt.Sprintf(`%v = "%v"`, fieldName, fieldName)
	case common.Boolean:
		return fmt.Sprintf(`%v = "%v"`, fieldName, "true")
	case common.Integer:
		return fmt.Sprintf(`%v = %v`, fieldName, "123")
	case common.StringList:
		if field.ItemType[service] == common.Integer {
			return fmt.Sprintf("%v = [%v]", fieldName, "1")
		}
		return fmt.Sprintf(`%v = ["%v"]`, fieldName, fieldName)
	case common.ObjectList:
		if len(field.ItemFields) > 0 {
			subFields := make([]string, 0)
			for n, f := range field.ItemFields {
				subFields = append(subFields, getTfConfigForFieldImpl(n, service, f))
			}
			subFieldsStr := strings.Join(subFields, "\n\t")
			return fmt.Sprintf("%v {\n\t%v\n}",
				fieldName, subFieldsStr)
		}
	}

	return ""
}

func getJsonConfigForField(fieldName, service string) string {
	if f, ok := common.GetConfigFieldsMap()[fieldName]; ok {
		return getJsonConfigForFieldImpl(fieldName, service, f)
	}
	return ""
}

func getJsonConfigForFieldImpl(fieldName, service string, field common.ConfigField) string {
	apiFieldName := fieldName
	if field.ApiField != "" {
		apiFieldName = field.ApiField
	}
	switch field.FieldValueType {
	case common.String:
		if field.ItemType != nil {
			if t, ok := field.ItemType[service]; ok {
				switch t {
				case common.Integer:
					return fmt.Sprintf(`"%v": %v`, apiFieldName, "123")
				case common.Float:
					return fmt.Sprintf(`"%v": %v`, apiFieldName, "123.4")
				}
			}
		}
		return fmt.Sprintf(`"%v": "%v"`, apiFieldName, fieldName)
	case common.Boolean:
		return fmt.Sprintf(`"%v": %v`, apiFieldName, "true")
	case common.Integer:
		return fmt.Sprintf(`"%v": %v`, apiFieldName, "123")
	case common.StringList:
		if field.ItemType[service] == common.Integer {
			return fmt.Sprintf(`"%v": [%v]`, apiFieldName, "1")
		}
		return fmt.Sprintf(`"%v": ["%v"]`, apiFieldName, fieldName)
	case common.ObjectList:
		if len(field.ItemFields) > 0 {
			subFields := make([]string, 0)
			for n, f := range field.ItemFields {
				subFields = append(subFields, getJsonConfigForFieldImpl(n, service, f))
			}
			subFieldsStr := strings.Join(subFields, ",\n\t")
			return fmt.Sprintf("\"%v\": [{\n\t%v\n}]",
				apiFieldName, subFieldsStr)
		}
	}

	return ""
}

func getTfDestinationSchema(service string) string {

	schemaFields := common.GetDestinationSchemaFields()
	if schemaFields[service]["schema"] {
		if schemaFields[service]["table"] {
			return fmt.Sprintf("\n\tname = \"%v\"\n\ttable = \"table\"\n", service)
		}
	} else {
		return fmt.Sprintf("\n\tprefix = \"%v\"\n", service)
	}
	return fmt.Sprintf("\n\tname = \"%v\"\n", service)
}

func getJsonSchemaValue(service string) string {
	if common.GetDestinationSchemaFields()[service]["schema"] {
		if common.GetDestinationSchemaFields()[service]["table"] {
			return fmt.Sprintf("%v.table", service)
		}
	}
	return service
}

func fetchFieldsBatchByService(fields []string) ([]string, []string, string) {
	if len(fields) > 0 {
		f := fields[0]
		if field, ok := common.GetConfigFieldsMap()[f]; ok {
			var service string
			if len(field.Description) == 0 {
				fmt.Printf("SKIP: field %v doesn't have definitions\n", f)
				return make([]string, 0), fields[1:], ""
			}
			fieldsCount := 0
			for s := range field.Description {
				fields, _ := common.GetFieldsForService(s)
				if len(fields) > fieldsCount {
					service = s
					fieldsCount = len(fields)
				}
			}
			fieldsToTest := []string{}
			restFields := []string{}
			serviceFields, _ := common.GetFieldsForService(service)
			for _, fName := range fields {
				if sf, ok := serviceFields[fName]; ok {
					if !sf.Readonly {
						fieldsToTest = append(fieldsToTest, fName)
					} else {
						fmt.Printf("SKIP: field %v - readonly", f)
					}
				} else {
					restFields = append(restFields, fName)
				}
			}
			return fieldsToTest, restFields, service
		} else {
			fmt.Printf("SKIP: field %v not defined in schema", f)
			return make([]string, 0), fields[1:], ""
		}
	}
	return make([]string, 0), make([]string, 0), ""
}

func getSortedFields() *[]string {
	if fields == nil || len(*fields) == 0 {
		fieldsMap := common.GetConfigFieldsMap()
		// Extract keys from map
		result := make([]string, 0, len(fieldsMap))
		for k, v := range fieldsMap {
			if !v.Readonly {
				result = append(result, k)
			}
		}

		// Sort keys
		sort.Strings(result)
		fields = &result
	}
	return fields
}

var fields *[]string

func TestResourceConnectorDynamicByServiceMapping(t *testing.T) {
	t.Skip("This test is for manual testing & debug for particular field")
	rf := []string{"account_key"}

	//restFields := &rf

	//_, _, service := fetchFieldsBatchByService(*restFields)

	service := "cosmos"

	if len(rf) > 0 {
		tfConfig := make([]string, 0)
		jsonConfig := make([]string, 0)
		for _, fname := range rf {
			tfConfig = append(tfConfig, getTfConfigForField(fname, service))
			jsonConfig = append(jsonConfig, getJsonConfigForField(fname, service))
		}
		tfc := strings.Join(tfConfig, "\n\t\t")
		jsonc := strings.Join(jsonConfig, ",\n\t\t")

		testResourceConnectorConfigConflictingFieldsMappingMock(t,
			service,
			getTfDestinationSchema(service),
			getJsonSchemaValue(service),
			tfc,
			jsonc,
		)
	}

}

func TestResourceConnectorDynamicMapping(t *testing.T) {
	restFields := getSortedFields()

	for len(*restFields) > 0 {
		stepFields, rest, service := fetchFieldsBatchByService(*restFields)

		fmt.Printf("Service %v | Fields left to test: %v | Step fields count: %v | Fields rest %v \n", service, len(rest)+len(stepFields), len(stepFields), len(rest))

		if debug {
			fmt.Printf("Testing fields for service %v : [\t\n%v]\n", service, strings.Join(stepFields, "\t\n "))
		}
		if len(stepFields) > 0 {
			tfConfig := make([]string, 0)
			jsonConfig := make([]string, 0)
			for _, fname := range stepFields {
				tfConfig = append(tfConfig, getTfConfigForField(fname, service))
				jsonConfig = append(jsonConfig, getJsonConfigForField(fname, service))
			}
			tfc := strings.Join(tfConfig, "\n\t\t")
			jsonc := strings.Join(jsonConfig, ",\n\t\t")

			testResourceConnectorConfigConflictingFieldsMappingMock(t,
				service,
				getTfDestinationSchema(service),
				getJsonSchemaValue(service),
				tfc,
				jsonc,
			)
		}
		if debug {
			fmt.Printf("Fields left to test: %v \n", len(rest))
		}
		restFields = &rest
	}
}

func testResourceConnectorConfigConflictingFieldsMappingMock(t *testing.T, service, destinationSchema, schema, tfConfig, jsonConfig string) {
	config := getTfConfigForConflictingFields(service, destinationSchema, tfConfig)

	if debug {
		fmt.Println("Final config: ")
		fmt.Println(config)
	}

	step1 := resource.TestStep{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			func(s *terraform.State) error {
				assertEqual(t, connectorConflictingMockPostHandler.Interactions, 1)
				assertEqual(t, connectorConflictingMockGetHandler.Interactions, 0)
				assertNotEmpty(t, connectorConflictingMappingMockData)
				return nil
			},
		),
	}

	resource.Test(
		t,
		resource.TestCase{
			PreCheck: func() {
				setupMockClientConnectorResourceConfigConflictingFieldsMapping(t, service, schema, jsonConfig)
			},
			ProtoV6ProviderFactories: ProtoV6ProviderFactories,
			CheckDestroy: func(s *terraform.State) error {
				assertEqual(t, connectorConflictingMockDelete.Interactions, 1)
				assertEmpty(t, connectorConflictingMappingMockData)
				return nil
			},

			Steps: []resource.TestStep{
				step1,
			},
		},
	)
}
