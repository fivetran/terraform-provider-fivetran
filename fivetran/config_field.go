package fivetran

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
