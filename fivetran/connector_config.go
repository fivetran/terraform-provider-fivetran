package fivetran

import (
	"reflect"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const MASKED_VALUE = "******"

func connectorSchemaAuth() *schema.Schema {
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
			// if valueMap, ok := valueArray[0].(map[string]interface{}); ok {
			// 	if valueMap["value"] == MASKED_VALUE && originConfig[responseProperty] != nil {
			// 		continue
			// 	}
			// }
			// if valueStringArray, ok := valueArray[0].([]string); ok {
			// 	if valueStringArray[0] == MASKED_VALUE {
			// 		continue
			// 	}
			// }
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

// connectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func getConnectorReadCustomConfig(resp *fivetran.ConnectorCustomDetailsResponse, currentConfigs *[]interface{}) []interface{} {
	configs := make([]interface{}, 1)

	config := make(map[string]interface{})
	originConfig := make(map[string]interface{})
	if currentConfigs != nil && len(*currentConfigs) > 0 {
		originConfig = (*currentConfigs)[0].(map[string]interface{})
	}

	originConfig = populateOriginConfigFromResponse(originConfig, resp.Data.Config)

	properties := getProperties()

	for property, propertySchema := range properties {

		if _, ok := sensitiveFields[property]; ok {
			if v, ok := originConfig[property].(string); ok {
				mapAddStr(config, property, v)
			}
			if v, ok := originConfig[property].([]interface{}); ok {
				mapAddXInterface(config, property, v)
			}
			continue
		}

		if propertySchema.Type == schema.TypeSet || propertySchema.Type == schema.TypeList {
			if values, ok := originConfig[property].([]string); ok {
				config[property] = xStrXInterface(values)
				continue
			}

			if interfaceValues, ok := originConfig[property].([]interface{}); ok && len(interfaceValues) > 0 {
				if _, ok := interfaceValues[0].(map[string]interface{}); ok {
					config[property] = interfaceValues
				} else {
					config[property] = xInterfaceStrXStr(interfaceValues)
				}
				continue
			}
		}
		if value, ok := originConfig[property].(string); ok && value != "" {
			valueType := propertySchema.Type
			switch valueType {
			case schema.TypeBool:
				config[property] = strToBool(value)
			case schema.TypeInt:
				config[property] = strToInt(value)
			default:
				config[property] = value
			}
		}
	}
	configs[0] = config

	return configs
}
