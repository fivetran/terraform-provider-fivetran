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

func connectorSchemaConfig(readonly bool) *schema.Schema {
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

func tryCopySensitiveStringValue(localConfig *map[string]interface{}, targetConfig, upstreamConfig map[string]interface{}, name string) {
	if localConfig == nil {
		tryCopyStringValue(targetConfig, upstreamConfig, name)
	} else {
		tryCopyStringValue(targetConfig, *localConfig, name)
	}
}

func tryCopySensitiveListValue(localConfig *map[string]interface{}, targetConfig, upstreamConfig map[string]interface{}, name string) {
	if localConfig != nil {
		mapAddXInterface(targetConfig, name, (*localConfig)[name].(*schema.Set).List())
	} else {
		tryCopyList(targetConfig, upstreamConfig, name)
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

	for k, v := range GetConfigFieldsMap() {
		readFieldValueCore(k, v, currentConfigMap, c, resp.Data.Config, service)
	}

	return []interface{}{c}
}

func readFieldValueCore(
	k string,
	v ConfigField,
	currentConfig *map[string]interface{},
	c map[string]interface{},
	upstream map[string]interface{},
	service string) {
	switch v.FieldValueType {
	case String:
		if v.Sensitive {
			tryCopySensitiveStringValue(currentConfig, c, upstream, k)
		} else {
			tryCopyStringValue(c, upstream, k)
		}
	case Integer:
		tryCopyIntegerValue(c, upstream, k)
	case Boolean:
		tryCopyBooleanValue(c, upstream, k)
	case StringList:
		if v.Sensitive {
			tryCopySensitiveListValue(currentConfig, c, upstream, k)
		} else {
			if t, ok := v.ItemType[service]; ok && t != String {
				if t == Integer {
					tryCopyIntegersList(c, upstream, k)
				}
			} else {
				tryCopyList(c, upstream, k)
			}
		}
	case ObjectList:
		var upstreamList = tryReadListValue(upstream, k)
		if upstreamList == nil || len(upstreamList) < 1 {
			mapAddXInterface(c, k, make([]interface{}, 0))
		} else {
			resultList := make([]interface{}, len(upstreamList))
			for i, elem := range upstreamList {
				upstreamElem := elem.(map[string]interface{})
				resultElem := make(map[string]interface{})
				localElem := getCorrespondingLocalElem(upstreamElem, currentConfig, k, v)
				for fn, fv := range v.ItemFields {
					readFieldValueCore(fn, fv, localElem, resultElem, upstreamElem, service)
				}
				resultList[i] = resultElem
			}
			mapAddXInterface(c, k, resultList)
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
	for k, v := range c {
		if field, ok := GetConfigFieldsMap()[k]; ok {
			updateConfigFieldImpl(k, field, v, configMap, service)
		}
	}
	return configMap
}

func updateConfigFieldImpl(name string, field ConfigField, v interface{}, configMap map[string]interface{}, service string) {
	switch field.FieldValueType {
	case String:
		{
			if v.(string) != "" {
				configMap[name] = v
			}
		}
	case Integer:
		{
			if v.(string) != "" {
				configMap[name] = strToInt(v.(string))
			}
		}
	case StringList:
		{
			if t, ok := field.ItemType[service]; ok && t != String {
				if t == Integer {
					configMap[name] = xInterfaceStrXIneger(v.(*schema.Set).List())
				}
			} else {
				configMap[name] = xInterfaceStrXStr(v.(*schema.Set).List())
			}
		}
	case Boolean:
		{
			if v.(string) != "" {
				configMap[name] = strToBool(v.(string))
			}
		}
	case ObjectList:
		{
			var list = v.(*schema.Set).List()
			result := make([]interface{}, len(list))
			for i, v := range list {
				vmap := v.(map[string]interface{})
				item := make(map[string]interface{})
				for subName, subField := range field.ItemFields {
					updateConfigFieldImpl(subName, subField, vmap[subName], item, service)
				}
				result[i] = item
			}
			configMap[name] = result
		}
	}
}
