package model

import (
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
