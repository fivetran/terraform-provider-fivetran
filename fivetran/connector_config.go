package fivetran

import (
	_ "embed"
	"fmt"
	"strings"

	"encoding/json"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type FieldValueType int64

/*
The following directive will do the following:
- process fivetran/open-api-spec.json file, fetch connector config fields and descriptions
- generate fivetran/fields-updated.json with meta-information for connector resource schema
- generate fivetran/config-changes file with a changelog

After calling go generate validate changelog, if everything is OK - replace existing fields.json with fields-update.json
*/

//go:generate go run ../utils/generate_connector_config.go

//go:embed fields.json
var fieldsJson []byte

func GetConfigFieldsMap() map[string]ConfigField {
	if len(configFields) == 0 {
		readFieldsFromJson(&configFields)
	}
	return configFields
}

func getMaxItems(readonly bool) int {
	if readonly {
		return 0
	}
	return 1
}

func readFieldsFromJson(target *map[string]ConfigField) {
	err := json.Unmarshal(fieldsJson, target)
	if err != nil {
		panic(err)
	}
	// handle and remove destination schema fields
	handleDestinationSchemaField("schema")
	handleDestinationSchemaField("table")
	handleDestinationSchemaField("schema_prefix")
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

var configFields = make(map[string]ConfigField)

var destinationSchemaFields = make(map[string]map[string]bool)

func GetDestinationSchemaFields() map[string]map[string]bool {
	return destinationSchemaFields
}

func getFieldSchema(isDataSourceSchema bool, field *ConfigField) *schema.Schema {
	result := &schema.Schema{
		Type:      schema.TypeString,
		Optional:  !isDataSourceSchema,
		Computed:  isDataSourceSchema || !field.Nullable,
		Sensitive: field.Sensitive,
	}

	if field.Readonly {
		if field.FieldValueType == StringList {
			result = &schema.Schema{Type: schema.TypeSet, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}}
		} else {
			result = &schema.Schema{Type: schema.TypeString, Computed: true}
		}
	} else {
		if field.FieldValueType == StringList {
			result = &schema.Schema{
				Type:      schema.TypeSet,
				Optional:  !isDataSourceSchema,
				Computed:  isDataSourceSchema,
				Sensitive: field.Sensitive,
				Elem:      &schema.Schema{Type: schema.TypeString}}
		} else if field.FieldValueType == ObjectList {
			var elemSchema = map[string]*schema.Schema{}

			for k, v := range field.ItemFields {
				elemSchema[k] = getFieldSchema(isDataSourceSchema, &v)
			}

			result = &schema.Schema{
				Type:     schema.TypeSet,
				Optional: !isDataSourceSchema,
				Computed: isDataSourceSchema,
				Elem: &schema.Resource{
					Schema: elemSchema,
				},
			}
		}
	}

	result.Description = buildDescription(field.Description)
	return result
}

func buildDescription(fieldDescription map[string]string) string {
	var result []string
	for service, description := range fieldDescription {
		result = append(result, fmt.Sprintf("\t- Service `%v`: %v", service, description))
	}
	return "Field usage depends on `service` value: \n" + strings.Join(result, "\n")
}

func getConnectorSchemaConfig(readonly bool) *schema.Schema {
	var schemaMap = map[string]*schema.Schema{}

	for k, v := range GetConfigFieldsMap() {
		schemaMap[k] = getFieldSchema(readonly, &v)
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: getMaxItems(readonly),
		Elem: &schema.Resource{
			Schema: schemaMap,
		},
	}
}

// connectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func connectorReadCustomConfig(resp *fivetran.ConnectorCustomDetailsResponse, currentConfig *[]interface{}, service string) []interface{} {
	c := make(map[string]interface{})
	var currentConfigMap *map[string]interface{} = nil

	if currentConfig != nil && len(*currentConfig) > 0 {
		vlocalConfigAsMap := (*currentConfig)[0].(map[string]interface{})
		currentConfigMap = &vlocalConfigAsMap
	}

	for k, v := range getServiceSpecificFields(service) {
		readFieldValueCore(k, v, currentConfigMap, c, resp.Data.Config, service)
	}

	return []interface{}{c}
}

func getServiceSpecificFields(service string) map[string]ConfigField {
	result := make(map[string]ConfigField)
	allFields := GetConfigFieldsMap()
	serviceSuffix := "_" + service

	for k, v := range GetConfigFieldsMap() {
		if v.ApiField == "" {
			// no service specific pair
			if _, ok := allFields[k+serviceSuffix]; !ok {
				result[k] = v
			}
		} else {
			// correct service
			if strings.HasSuffix(k, serviceSuffix) {
				result[k] = v
			}
		}
	}

	return result
}

func readFieldValueCore(
	fieldName string, field ConfigField, localConfig *map[string]interface{},
	c map[string]interface{}, upstream map[string]interface{}, service string) {
	switch field.FieldValueType {
	case String:
		if field.Sensitive {
			copySensitiveStringValue(localConfig, c, upstream, fieldName, field.ApiField)
		} else {
			copyStringValue(c, upstream, fieldName, field.ApiField)
		}
	case Integer:
		copyIntegerValue(c, upstream, fieldName, field.ApiField)
	case Boolean:
		copyBooleanValue(c, upstream, fieldName, field.ApiField)
	case StringList:
		if field.Sensitive {
			copySensitiveListValue(localConfig, c, upstream, fieldName, field.ApiField)
		} else {
			if t, ok := field.ItemType[service]; ok && t != String {
				if t == Integer {
					copyIntegersList(c, upstream, fieldName, field.ApiField)
				}
			} else {
				copyList(c, upstream, fieldName, field.ApiField)
			}
		}
	case ObjectList:
		upstreamFieldName := fieldName
		if field.ApiField != "" {
			upstreamFieldName = field.ApiField
		}
		upstreamList := tryReadListValue(upstream, upstreamFieldName)
		if upstreamList == nil || len(upstreamList) < 1 {
			mapAddXInterface(c, fieldName, make([]interface{}, 0))
		} else {
			resultList := make([]interface{}, len(upstreamList))
			for i, elem := range upstreamList {
				upstreamElem := elem.(map[string]interface{})
				resultElem := make(map[string]interface{})
				localElem := getCorrespondingLocalElem(upstreamElem, localConfig, fieldName, field)
				for fn, fv := range field.ItemFields {
					readFieldValueCore(fn, fv, localElem, resultElem, upstreamElem, service)
				}
				resultList[i] = resultElem
			}
			mapAddXInterface(c, fieldName, resultList)
		}
	}
}

func getCorrespondingLocalElem(upstreamElem map[string]interface{}, currentConfig *map[string]interface{}, k string, v ConfigField) *map[string]interface{} {
	if v.ItemKeyField == "" {
		return nil
	}

	subKeyValue := tryReadValue(upstreamElem, v.ItemKeyField)

	if currentConfig != nil && subKeyValue != nil {
		targetList := (*currentConfig)[k].(*schema.Set).List()

		var filterFunc = func(elem interface{}) bool {
			return elem.(map[string]interface{})[v.ItemKeyField].(string) == subKeyValue
		}
		found := filterList(targetList, filterFunc)
		if found != nil {
			foundAsMap := (*found).(map[string]interface{})
			return &foundAsMap
		}
	}
	return nil
}

func connectorUpdateCustomConfig(c map[string]interface{}, service string) map[string]interface{} {
	configMap := make(map[string]interface{})
	serviceFields := getServiceSpecificFields(service)
	for k, v := range c {
		if field, ok := serviceFields[k]; ok {
			updateConfigFieldImpl(k, field, v, configMap, service)
		}
	}
	return configMap
}

func updateConfigFieldImpl(name string, field ConfigField, value interface{}, configMap map[string]interface{}, service string) {
	upstreamFieldName := name
	if field.ApiField != "" {
		upstreamFieldName = field.ApiField
	}
	switch field.FieldValueType {
	case String:
		{
			if value.(string) != "" {
				configMap[upstreamFieldName] = value
			}
		}
	case Integer:
		{
			if value.(string) != "" {
				configMap[upstreamFieldName] = strToInt(value.(string))
			}
		}
	case StringList:
		{
			if t, ok := field.ItemType[service]; ok && t != String {
				if t == Integer {
					configMap[upstreamFieldName] = xInterfaceStrXIneger(value.(*schema.Set).List())
				}
			} else {
				configMap[upstreamFieldName] = xInterfaceStrXStr(value.(*schema.Set).List())
			}
		}
	case Boolean:
		{
			if value.(string) != "" {
				configMap[upstreamFieldName] = strToBool(value.(string))
			}
		}
	case ObjectList:
		{
			var list = value.(*schema.Set).List()
			result := make([]interface{}, len(list))
			for i, v := range list {
				vmap := v.(map[string]interface{})
				item := make(map[string]interface{})
				for subName, subField := range field.ItemFields {
					updateConfigFieldImpl(subName, subField, vmap[subName], item, service)
				}
				result[i] = item
			}
			configMap[upstreamFieldName] = result
		}
	}
}
