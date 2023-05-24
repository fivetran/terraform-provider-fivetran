package fivetran

import (
	"fmt"

	"github.com/Jeffail/gabs/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const SCHEMAS_PATH = "schemas."
const PROPERTIES_PATH = ".properties.config.properties"
const SCHEMAS_FILE_PATH = "/Users/lukadevic/Fivetran/terraform-provider-fivetran/fivetran/schemas.json"
const SERVICES_FILE_PATH = "/Users/lukadevic/Fivetran/terraform-provider-fivetran/fivetran/services.json"

const OBJECT_PROPERTY_TYPE = "object"
const INT_PROPERTY_TYPE = "integer"
const BOOL_PROPERTY_TYPE = "boolean"
const ARRAY_PROPERTY_TYPE = "array"
const STRING_PROPERTY_TYPE = "string"

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

func getSchemaAndProperties(path string) map[string]*schema.Schema {
	shemasJson, err := gabs.ParseJSONFile(SCHEMAS_FILE_PATH)
	if err != nil {
		panic(err)
	}

	nodesMap := shemasJson.Path(path).ChildrenMap()

	properties := getProperties(nodesMap)

	return properties
}

func getProperties(nodesMap map[string]*gabs.Container) map[string]*schema.Schema {
	properties := make(map[string]*schema.Schema)

	for key, node := range nodesMap {
		nodeSchema := &schema.Schema{
			Type:     schema.TypeString,
			Computed: true}

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
		Computed: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		}}

	if itemType == OBJECT_PROPERTY_TYPE && len(childrenMap) > 0 {
		childrenSchemaMap := getProperties(childrenMap)

		arraySchema.Elem = &schema.Resource{
			Schema: childrenSchemaMap,
		}
	}

	return arraySchema
}

func getCSchemaAndProperties(path string) map[string]*schema.Schema {
	shemasJson, err := gabs.ParseJSONFile(SCHEMAS_FILE_PATH)
	if err != nil {
		panic(err)
	}

	nodesMap := shemasJson.Path(path).ChildrenMap()

	properties := getCProperties(nodesMap)

	return properties
}

func getCProperties(nodesMap map[string]*gabs.Container) map[string]*schema.Schema {
	properties := make(map[string]*schema.Schema)

	for key, node := range nodesMap {
		if key == "advertisers_id" {
			fmt.Printf("this property is now:%v", key)
		}
		nodeSchema := &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Computed: true}

		nodeType := node.Search("type").Data()

		switch nodeType {
		case INT_PROPERTY_TYPE:
			nodeSchema.Type = schema.TypeInt
		case BOOL_PROPERTY_TYPE:
			nodeSchema.Type = schema.TypeBool
		case ARRAY_PROPERTY_TYPE:
			nodeSchema = getCArrayPropertySchema(node)
		}
		properties[key] = nodeSchema
	}

	return properties
}

func getCArrayPropertySchema(node *gabs.Container) *schema.Schema {
	itemType := node.Path("items.type").Data()

	childrenMap := node.Path("items.properties").ChildrenMap()

	arraySchema := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		}}

	// "breakdowns": {
	// 	Type: schema.TypeSet,
	// 	 Optional: true,
	// 	  Elem: &schema.Schema
	// 	  {Type: schema.TypeString}},

	// "apps":   {
	// 	Type: schema.TypeSet,
	// 	 Optional: true,
	// 	  Elem: &schema.Schema{
	// 		Type: schema.TypeString}},

	if itemType == OBJECT_PROPERTY_TYPE && len(childrenMap) > 0 {
		childrenSchemaMap := getCProperties(childrenMap)

		arraySchema.Elem = &schema.Resource{
			Schema: childrenSchemaMap,
		}
	}

	// if itemType == INT_PROPERTY_TYPE {
	// 	arraySchema.Elem = &schema.Schema{
	// 		Type: schema.TypeInt,
	// 	}
	// }

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
