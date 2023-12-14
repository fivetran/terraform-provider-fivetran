package common

import (
	_ "embed"
	"encoding/json"
)

type ConfigField struct {
	Readonly       bool                      `json:"readonly"`
	Sensitive      bool                      `json:"sensitive"`
	Nullable       bool                      `json:"nullable"`
	FieldValueType FieldValueType            `json:"type"`
	ItemFields     map[string]ConfigField    `json:"fields"`
	ItemKeyField   string                    `json:"key_field"`
	ItemType       map[string]FieldValueType `json:"item_type"`
	Description    map[string]string         `json:"description"`
	ApiField       string                    `json:"api_field"`
}

func NewconfigField() ConfigField {
	field := ConfigField{}
	field.FieldValueType = String
	field.Readonly = false
	field.Sensitive = false
	field.Nullable = true
	field.Description = make(map[string]string)
	field.ApiField = ""
	field.ItemKeyField = ""
	field.ItemType = make(map[string]FieldValueType)
	return field
}

//go:embed fields.json
var configFieldsJson []byte

//go:embed auth-fields.json
var authFieldsJson []byte

var configFields = make(map[string]ConfigField)

var authFields = make(map[string]ConfigField)

var configFieldsByService = make(map[string]map[string]ConfigField)
var authFieldsByService = make(map[string]map[string]ConfigField)

var destinationSchemaFields = make(map[string]map[string]bool)

func GetFieldsForService(service string) map[string]ConfigField {
	if len(configFieldsByService) == 0 {
		readFieldsFromJson(&configFields)
	}
	if r, ok := configFieldsByService[service]; ok {
		return r
	}
	panic("Unknown service" + service)
}

func GetAuthFieldsForService(service string) map[string]ConfigField {
	if len(authFieldsByService) == 0 {
		GetAuthFieldsMap()
	}
	if r, ok := authFieldsByService[service]; ok {
		return r
	}
	return map[string]ConfigField{}
}

func GetDestinationSchemaFields() map[string]map[string]bool {
	if len(destinationSchemaFields) == 0 {
		readFieldsFromJson(&configFields)
	}
	return destinationSchemaFields
}

func GetAuthFieldsMap() map[string]ConfigField {
	if len(authFields) == 0 {
		readAuthFieldsFromJson(&authFields)
	}

	authFieldsByService = make(map[string]map[string]ConfigField)
	for k, v := range authFields {
		for service := range v.Description {
			if _, ok := authFieldsByService[service]; !ok {
				authFieldsByService[service] = make(map[string]ConfigField)
			}
			authFieldsByService[service][k] = v
		}
	}
	return authFields
}

func GetConfigFieldsMap() map[string]ConfigField {
	if len(configFields) == 0 {
		readFieldsFromJson(&configFields)
	}
	return configFields
}

func readAuthFieldsFromJson(target *map[string]ConfigField) {
	if err := json.Unmarshal(authFieldsJson, target); err != nil {
		panic(err)
	}
}

func readFieldsFromJson(target *map[string]ConfigField) {
	err := json.Unmarshal(configFieldsJson, target)
	if err != nil {
		panic(err)
	}
	// handle and remove destination schema fields
	handleDestinationSchemaField("schema")
	handleDestinationSchemaField("table")
	handleDestinationSchemaField("schema_prefix")

	// sort rest fields by service
	fillFieldsByService()
}

func handleDestinationSchemaField(fieldName string) {
	if schema, ok := configFields[fieldName]; ok {
		for k := range schema.Description {
			if _, ok := destinationSchemaFields[k]; !ok {
				destinationSchemaFields[k] = make(map[string]bool)
			}
			destinationSchemaFields[k][fieldName] = true
		}
		delete(configFields, fieldName)
	}
}

func fillFieldsByService() {
	configFieldsByService = make(map[string]map[string]ConfigField)
	for k, v := range configFields {
		for service := range v.Description {
			if _, ok := configFieldsByService[service]; !ok {
				configFieldsByService[service] = make(map[string]ConfigField)
			}
			configFieldsByService[service][k] = v
		}
	}
}
