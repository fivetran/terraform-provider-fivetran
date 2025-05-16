package common

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type ConfigField struct {
	Readonly            bool                      `json:"readonly"`
	Sensitive           bool                      `json:"sensitive"`
	Nullable            bool                      `json:"nullable"`
	FieldValueType      FieldValueType            `json:"type"`
	ItemFields          map[string]ConfigField    `json:"fields"`
	ItemKeyField        string                    `json:"key_field"`
	ItemType            map[string]FieldValueType `json:"item_type"`
	Description         map[string]string         `json:"description"`
	SensitiveExclusions map[string]bool           `json:"sensitive_exclusions,omitempty"`
	ApiField            string                    `json:"api_field"`
}

func (c ConfigField) GetIsSensitiveForSchema() bool {
	return c.Sensitive || (c.SensitiveExclusions != nil && len(c.SensitiveExclusions) > 0)
}

func (c ConfigField) GetIsSensitive(service string) bool {
	if !c.Sensitive {
		if c.SensitiveExclusions != nil {
			_, ok := c.SensitiveExclusions[service]
			return ok
		}
		return false
	}
	return true
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
	field.SensitiveExclusions = nil
	return field
}

var (
	//go:embed fields.json
	configFieldsJson []byte

	//go:embed auth-fields.json
	authFieldsJson []byte

	//go:embed destination-fields.json
	destinationFieldsJson []byte

	configFields      = make(map[string]ConfigField)
	authFields        = make(map[string]ConfigField)
	destinationFields = make(map[string]ConfigField)

	configFieldsByService      = make(map[string]map[string]ConfigField)
	authFieldsByService        = make(map[string]map[string]ConfigField)
	destinationFieldsByService = make(map[string]map[string]ConfigField)

	destinationSchemaFields = make(map[string]map[string]bool)
)

func GetFieldsForService(service string) (map[string]ConfigField, error) {
	if len(configFieldsByService) == 0 {
		panic("Fields for config are not loaded")
	}
	if r, ok := configFieldsByService[service]; ok {
		return r, nil
	}
	if _, ok := destinationSchemaFields[service]; ok {
		return map[string]ConfigField{}, nil
	}
	return nil, fmt.Errorf("Unknown service: %v\n It seems like `%v` service is not yet supported in this provider version. \nPlease update to latest or wait for next release (if you are using latest already).", service, service)
}

func GetAuthFieldsForService(service string) map[string]ConfigField {
	if len(authFieldsByService) == 0 {
		panic("Fields for auth are not loaded")
	}
	if r, ok := authFieldsByService[service]; ok {
		return r
	}
	return map[string]ConfigField{}
}

func GetDestinationFieldsForService(service string) map[string]ConfigField {
	if len(destinationFieldsByService) == 0 {
		panic("Fields for destination config are not loaded")
	}
	if r, ok := destinationFieldsByService[service]; ok {
		return r
	}
	return map[string]ConfigField{}
}

func GetDestinationSchemaFields() map[string]map[string]bool {
	if len(destinationSchemaFields) == 0 {
		panic("Fields for config are not loaded")
	}
	return destinationSchemaFields
}

func GetAuthFieldsMap() map[string]ConfigField {
	if len(authFields) == 0 {
		panic("Fields for auth are not loaded")
	}
	return authFields
}

func GetConfigFieldsMap() map[string]ConfigField {
	if len(configFields) == 0 {
		panic("Fields for config are not loaded")
	}
	return configFields
}

func GetDestinationFieldsMap() map[string]ConfigField {
	if len(destinationFields) == 0 {
		panic("Fields for destination config are not loaded")
	}
	return destinationFields
}

func LoadAuthFieldsMap() {
	if len(authFields) == 0 {
		readAuthFieldsFromJson(&authFields)
	}
	fillAuthFieldsByService()
}

func LoadConfigFieldsMap() {
	if len(configFields) == 0 {
		readFieldsFromJson(&configFields)
	}
	fillFieldsByService()
}

func LocaDestinationFieldsMap() {
	if len(destinationFields) == 0 {
		readDestinationFieldsFromJson(&destinationFields)
	}
	fillDestinationFieldsByService()
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
	handleDestinationSchemaField("table_group_name")
}

func readDestinationFieldsFromJson(target *map[string]ConfigField) {
	err := json.Unmarshal(destinationFieldsJson, target)
	if err != nil {
		panic(err)
	}
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

func fillDestinationFieldsByService() {
	destinationFieldsByService = breakByService(destinationFields)
}

func fillFieldsByService() {
	configFieldsByService = breakByService(configFields)
}

func fillAuthFieldsByService() {
	authFieldsByService = breakByService(authFields)
}

func breakByService(source map[string]ConfigField) map[string]map[string]ConfigField {
	result := make(map[string]map[string]ConfigField)
	for k, v := range source {
		for service := range v.Description {
			if _, ok := result[service]; !ok {
				result[service] = make(map[string]ConfigField)
			}
			result[service][k] = v
		}
	}
	return result
}
