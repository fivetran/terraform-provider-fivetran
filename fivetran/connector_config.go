package fivetran

import (
	"reflect"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const MASKED_VALUE = "******"

func getConnectorSchemaAuth() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"client_access": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"client_id": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"client_secret": {
								Type:      schema.TypeString,
								Optional:  true,
								Sensitive: true,
							},
							"user_agent": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"developer_token": {
								Type:      schema.TypeString,
								Optional:  true,
								Sensitive: true,
							},
						},
					},
				},
				"refresh_token": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
				"access_token": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
				"realm_id": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
			},
		},
	}
}

// connectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func getConnectorReadCustomConfig(resp *fivetran.ConnectorCustomDetailsResponse, currentConfigs *[]interface{}) []interface{} {
	config := make(map[string]interface{})
	originConfig := make(map[string]interface{})
	if currentConfigs != nil && len(*currentConfigs) > 0 {
		originConfig = (*currentConfigs)[0].(map[string]interface{})
	}

	originConfig = populateOriginConfigFromResponse(originConfig, resp.Data.Config)

	fields := getFields()

	for fieldName, fieldSchema := range fields {
		if fieldSchema.Type == schema.TypeSet || fieldSchema.Type == schema.TypeList {
			if values, ok := originConfig[fieldName].([]string); ok {
				config[fieldName] = xStrXInterface(values)
				continue
			}

			if interfaceValues, ok := originConfig[fieldName].([]interface{}); ok && len(interfaceValues) > 0 {
				if _, ok := interfaceValues[0].(map[string]interface{}); ok {
					config[fieldName] = interfaceValues
				} else {
					config[fieldName] = xInterfaceStrXStr(interfaceValues)
				}
				continue
			}
		}
		if value, ok := originConfig[fieldName].(string); ok && value != "" {
			switch fieldSchema.Type {
			case schema.TypeBool:
				config[fieldName] = strToBool(value)
			case schema.TypeInt:
				config[fieldName] = strToInt(value)
			default:
				config[fieldName] = value
			}
		}
	}
	configs := make([]interface{}, 1)
	configs[0] = config

	return configs
}

func populateOriginConfigFromResponse(originConfig map[string]interface{}, responseConfig map[string]interface{}) map[string]interface{} {
	for responseProperty, value := range responseConfig {
		if responseProperty == "project_credentials" && originConfig[responseProperty] != nil {
			// Hack for project_credentials property
			continue
		}

		reflectedValue := reflect.ValueOf(value)
		if reflectedValue.Kind() == reflect.Slice && reflect.TypeOf(value).Elem().Kind() != reflect.String {
			var valueArray []interface{}
			for i := 0; i < reflectedValue.Len(); i++ {
				valueArray = append(valueArray, reflectedValue.Index(i).Interface())
			}
			if isMaskedValue(valueArray, originConfig, responseProperty) {
				continue
			}
			originConfig[responseProperty] = valueArray
			continue
		}
		if value != MASKED_VALUE || originConfig[responseProperty] == nil {
			originConfig[responseProperty] = value
		}
	}
	return originConfig
}

func isMaskedValue(valueArray []interface{}, originConfig map[string]interface{}, responseProperty string) bool {
	if valueMap, ok := valueArray[0].(map[string]interface{}); ok {
		if valueMap["value"] == MASKED_VALUE && originConfig[responseProperty] != nil {
			return true
		}
	}
	if valueStringArray, ok := valueArray[0].([]string); ok {
		if valueStringArray[0] == MASKED_VALUE {
			return true
		}
	}
	return false
}
