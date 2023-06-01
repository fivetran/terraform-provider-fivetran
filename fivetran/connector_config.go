package fivetran

import (
	"fmt"
	"reflect"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type FieldValueType int64
type FieldType int64

const (
	String     FieldValueType = 0
	Integer    FieldValueType = 1
	Boolean    FieldValueType = 2
	StringList FieldValueType = 3
	ObjectList FieldValueType = 4
)

type configField struct {
	readonly       bool
	sensitive      bool
	nullable       bool
	fieldValueType FieldValueType
	itemFields     map[string]configField
	itemKeyField   string
}

func NewconfigField() configField {
	field := configField{}
	field.fieldValueType = String
	field.readonly = false
	field.sensitive = false
	field.nullable = true
	return field
}

var sensitiveFields = map[string]bool{
	"oauth_token":        true,
	"oauth_token_secret": true,
	"consumer_key":       true,
	"client_secret":      true,
	"private_key":        true,
	"s3role_arn":         true,
	"ftp_password":       true,
	"sftp_password":      true,
	"api_key":            true,
	"role_arn":           true,
	"password":           true,
	"secret_key":         true,
	"pem_certificate":    true,
	"access_token":       true,
	"api_secret":         true,
	"api_access_token":   true,
	"secret":             true,
	"consumer_secret":    true,
	"secrets":            true,
	"api_token":          true,
	"encryption_key":     true,
	"pat":                true,
	"function_trigger":   true,
	"token_key":          true,
	"token_secret":       true,
	"agent_password":     true,
	"asm_password":       true,
	"login_password":     true,
}

func getConnectorSchemaConfig(readonly bool) *schema.Schema {
	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		oasProperties := getSchemaAndProperties(path)
		for key, value := range oasProperties {
			if existingValue, ok := properties[key]; ok {
				if existingValue.Type == schema.TypeList {
					if _, ok := existingValue.Elem.(map[string]*schema.Schema); ok {
						continue
					}
					value = updateExistingValue(existingValue, value)
				}
			}
			properties[key] = value
		}
	}

	return &schema.Schema{Type: schema.TypeList, Optional: true, Computed: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: properties,
		},
	}
}

func connectorSchemaAuth() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Optional: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"client_access": {Type: schema.TypeList, Optional: true, MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"client_id":       {Type: schema.TypeString, Optional: true},
							"client_secret":   {Type: schema.TypeString, Optional: true, Sensitive: true},
							"user_agent":      {Type: schema.TypeString, Optional: true},
							"developer_token": {Type: schema.TypeString, Optional: true, Sensitive: true},
						},
					},
				},
				"refresh_token": {Type: schema.TypeString, Optional: true, Sensitive: true},
				"access_token":  {Type: schema.TypeString, Optional: true, Sensitive: true},
				"realm_id":      {Type: schema.TypeString, Optional: true, Sensitive: true},
			},
		},
	}
}

// connectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func getConnectorReadCustomConfig(resp *fivetran.ConnectorCustomDetailsResponse, currentConfig *[]interface{}) []interface{} {
	configArray := make([]interface{}, 1)

	configResult := make(map[string]interface{})
	responseConfig := make(map[string]interface{})
	if currentConfig != nil && len(*currentConfig) > 0 {
		responseConfig = (*currentConfig)[0].(map[string]interface{})
	}

	responseConfigFromStruct := resp.Data.Config
	for responseProperty, value := range responseConfigFromStruct {
		if responseProperty == "project_credentials" && responseConfig[responseProperty] != nil {
			// Hack for project_credentials property
			continue
		}

		if responseProperty == "consumer_key" {
			fmt.Printf("consumer_key")
		}
		reflectedValue := reflect.ValueOf(value)
		if reflectedValue.Kind() == reflect.Slice && reflect.TypeOf(value).Elem().Kind() != reflect.String {
			var valueArray []interface{}
			for i := 0; i < reflectedValue.Len(); i++ {
				valueArray = append(valueArray, reflectedValue.Index(i).Interface())
			}

			childPropertiesFromStruct := valueArray[0]
			valueArray[0] = childPropertiesFromStruct
			if responseProperty == "secrets_list" {
				fmt.Printf("now")
			}

			if value1, ok := valueArray[0].(map[string]interface{}); ok {
				if value1["value"] == "******" && responseConfig[responseProperty] != nil {
					continue
				}
			}
			if value2, ok := valueArray[0].([]string); ok {

				if value2[0] == "******" {
					continue
				}
			}

			responseConfig[responseProperty] = valueArray
			continue

		}
		if value != "******" {
			responseConfig[responseProperty] = value
		}
		if value == "******" && responseConfig[responseProperty] == nil {
			responseConfig[responseProperty] = value
		}
	}

	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		newProperties := getSchemaAndProperties(path)
		for newPropertyKey, newPropertySchema := range newProperties {
			properties[newPropertyKey] = newPropertySchema
		}
	}

	for property, propertySchema := range properties {

		if _, ok := sensitiveFields[property]; ok {
			if v, ok := responseConfig[property].(string); ok {
				mapAddStr(configResult, property, v)
			}
			if v, ok := responseConfig[property].([]interface{}); ok {
				mapAddXInterface(configResult, property, v)
			}
			continue
		}

		if propertySchema.Type == schema.TypeSet || propertySchema.Type == schema.TypeList {
			if values, ok := responseConfig[property].([]string); ok {
				configResult[property] = xStrXInterface(values)
				continue
			}

			if interfaceValues, ok := responseConfig[property].([]interface{}); ok && len(interfaceValues) > 0 {
				if _, ok := interfaceValues[0].(map[string]interface{}); ok {
					configResult[property] = interfaceValues
				} else {
					configResult[property] = xInterfaceStrXStr(interfaceValues)
				}
				continue
			}
		}
		if value, ok := responseConfig[property].(string); ok && value != "" {
			valueType := propertySchema.Type
			switch valueType {
			case schema.TypeBool:
				configResult[property] = strToBool(value)
			case schema.TypeInt:
				configResult[property] = strToInt(value)
			default:
				configResult[property] = value
			}
		}
	}
	configArray[0] = configResult

	return configArray
}
