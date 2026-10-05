package model

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/fivetran/terraform-provider-fivetran/fivetran/framework/core/fivetrantypes"
)

func TestConfiguredPrimaryKeysFromSchemasJson(t *testing.T) {
	d := ConnectorSchemaResourceModel{
		Schema:  (&ConnectorSchemaResourceModel{}).getNullSchema(),
		Schemas: (&ConnectorSchemaResourceModel{}).getNullSchemas(),
		SchemasRaw: fivetrantypes.NewJsonSchemaValue(`{
			"public": {
				"tables": {
					"users": {
						"columns": {
							"id":   {"is_primary_key": true},
							"name": {"enabled": true}
						}
					}
				}
			}
		}`),
	}

	expected := map[string]map[string]map[string]bool{"public": {"users": {"id": true}}}
	if actual := d.ConfiguredPrimaryKeys(); !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func TestSchemasRawValueKeepsOnlyConfiguredPrimaryKeys(t *testing.T) {
	d := ConnectorSchemaResourceModel{
		SchemasRaw: fivetrantypes.NewJsonSchemaValue(`{
			"public": {"tables": {"users": {"columns": {
				"id":   {"enabled": true, "is_primary_key": true},
				"name": {"enabled": true}
			}}}}
		}`),
	}
	upstream := []interface{}{
		map[string]interface{}{
			"name":    "public",
			"enabled": "true",
			"table": []interface{}{
				map[string]interface{}{
					"name":    "users",
					"enabled": "true",
					"column": []interface{}{
						map[string]interface{}{"name": "id", "enabled": "true", "is_primary_key": "true"},
						map[string]interface{}{"name": "name", "enabled": "true", "is_primary_key": "false"},
					},
				},
			},
		},
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(d.getSchemasRawValue(upstream)), &raw); err != nil {
		t.Fatal(err)
	}
	columns := raw["public"].(map[string]interface{})["tables"].(map[string]interface{})["users"].(map[string]interface{})["columns"].(map[string]interface{})
	if _, ok := columns["id"].(map[string]interface{})["is_primary_key"]; !ok {
		t.Errorf("expected configured is_primary_key to be kept for `id`")
	}
	if _, ok := columns["name"].(map[string]interface{})["is_primary_key"]; ok {
		t.Errorf("expected unconfigured is_primary_key to be removed for `name`")
	}
}

func TestConfiguredPrimaryKeysFromSchemasJsonStringValues(t *testing.T) {
	d := ConnectorSchemaResourceModel{
		Schema:     (&ConnectorSchemaResourceModel{}).getNullSchema(),
		Schemas:    (&ConnectorSchemaResourceModel{}).getNullSchemas(),
		SchemasRaw: fivetrantypes.NewJsonSchemaValue(`{"public": {"tables": {"users": {"columns": {"id": {"is_primary_key": "true"}}}}}}`),
	}

	expected := map[string]map[string]map[string]bool{"public": {"users": {"id": true}}}
	if actual := d.ConfiguredPrimaryKeys(); !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
