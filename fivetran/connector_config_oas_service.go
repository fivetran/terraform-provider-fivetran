package fivetran

import (
	"github.com/Jeffail/gabs/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const SCHEMAS_PATH = "schemas."
const PROPERTIES_PATH = ".properties.config.properties"
const SCHEMAS_FILE_PATH = "/Users/lukadevic/Fivetran/terraform-provider-fivetran/fivetran/schemas.json"
const SERVICES_FILE_PATH = "/Users/lukadevic/Fivetran/terraform-provider-fivetran/fivetran/services.json"

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

func getConnectorSchemaConfig() *schema.Schema {
	properties := getProperties()

	return &schema.Schema{Type: schema.TypeList, Optional: true, Computed: true, MaxItems: 1,
		Elem: &schema.Resource{
			Schema: properties,
		},
	}
}

func getProperties() map[string]*schema.Schema {
	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		propertiesOasSchema := getOasSchema(path)
		oasProperties := getOasProperties(propertiesOasSchema)
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
	return properties
}

func getAvailableServiceIds() []string {
	servicesJson, err := gabs.ParseJSONFile(SERVICES_FILE_PATH)
	if err != nil {
		panic(err)
	}

	var services []string

	for serviceKey := range servicesJson.S("services").ChildrenMap() {
		services = append(services, serviceKey+"_config_V1")
	}

	return services
}

func getOasSchema(path string) map[string]*gabs.Container {
	shemasJson, err := gabs.ParseJSONFile(SCHEMAS_FILE_PATH)
	if err != nil {
		panic(err)
	}

	return shemasJson.Path(path).ChildrenMap()
}

func getOasProperties(nodesMap map[string]*gabs.Container) map[string]*schema.Schema {
	properties := make(map[string]*schema.Schema)

	for key, node := range nodesMap {
		if _, ok := sensitiveFields[key]; ok {
			properties[key] = &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true}
			continue
		}

		nodeSchema := &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		}

		nodeType := node.Search("type").Data()

		switch nodeType {
		case INT_PROPERTY_TYPE:
			nodeSchema.Type = schema.TypeInt
		case BOOL_PROPERTY_TYPE:
			nodeSchema.Type = schema.TypeBool
		case ARRAY_PROPERTY_TYPE:
			nodeSchema = getArrayPropertySchema(node)
		}
		properties[key] = nodeSchema
	}

	return properties
}

func getArrayPropertySchema(node *gabs.Container) *schema.Schema {
	itemType := node.Path("items.type").Data()

	childrenMap := node.Path("items.properties").ChildrenMap()

	arraySchema := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		}}

	if itemType == OBJECT_PROPERTY_TYPE && len(childrenMap) > 0 {
		childrenSchemaMap := getOasProperties(childrenMap)

		arraySchema.Elem = &schema.Resource{
			Schema: childrenSchemaMap,
		}
	}

	return arraySchema
}

func updateExistingValue(existingValue *schema.Schema, newValue *schema.Schema) *schema.Schema {
	if existingSchemaResourceValue, ok := existingValue.Elem.(*schema.Resource); ok {
		if newSchemaResourceValue, ok := newValue.Elem.(*schema.Resource); ok {
			for newSchemaResourceKey, newSchemaResourceValue := range newSchemaResourceValue.Schema {
				existingSchemaResourceValue.Schema[newSchemaResourceKey] = newSchemaResourceValue
			}
			existingValue.Elem = existingSchemaResourceValue
		}
	}
	return existingValue
}
