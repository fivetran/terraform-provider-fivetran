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
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Your application client access fields.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"client_id": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "`Client ID` of your client application.",
							},
							"client_secret": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "`Client secret` of your client application.",
							},
							"user_agent": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Your company's name in your client application.",
							},
							"developer_token": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "Your approved `Developer token` to connect to the API.",
							},
						},
					},
				},
				"refresh_token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The long-lived `Refresh token` along with the `client_id` and `client_secret` parameters carry the information necessary to get a new access token for API resources.",
				},
				"access_token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The `Access Token` carries the information necessary for API resources to fetch data.",
				},
				"realm_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "`Realm ID` of your application.",
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

	fields := getFields(false)

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
		if originConfig[responseProperty] != nil {
			if valueSet, ok := originConfig[responseProperty].(*schema.Set); ok {
				originConfig[responseProperty] = valueSet.List()
				continue
			}
		}

		reflectedValue := reflect.ValueOf(value)
		if reflectedValue.Kind() == reflect.Slice && reflect.TypeOf(value).Elem().Kind() != reflect.String {
			var valueArray []interface{}
			for i := 0; i < reflectedValue.Len(); i++ {
				valueArray = append(valueArray, reflectedValue.Index(i).Interface())
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
