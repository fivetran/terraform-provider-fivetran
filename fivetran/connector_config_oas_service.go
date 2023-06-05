package fivetran

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Jeffail/gabs/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/icza/dyno"
	"gopkg.in/yaml.v2"
)

const SCHEMAS_PATH = "components.schemas."
const PROPERTIES_PATH = ".properties.config.properties"
const SERVICES_FILE_PATH = "/services.json"
const SCHEMAS_FILE_PATH = "/oas.yaml"

const OBJECT_FIELD = "object"
const INT_FIELD = "integer"
const BOOL_FIELD = "boolean"
const ARRAY_FIELD = "array"

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
	fields := getFields()

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: fields,
		},
	}
}

func getFields() map[string]*schema.Schema {
	services := getAvailableServiceIds()

	fields := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		serviceSchema := getServiceSchema(path)
		serviceFields := createFields(serviceSchema)
		for property, value := range serviceFields {
			if existingValue, ok := fields[property]; ok {
				if existingValue.Type == schema.TypeList {
					if _, ok := existingValue.Elem.(map[string]*schema.Schema); ok {
						continue
					}
					value = updateExistingValue(existingValue, value)
				}
			}
			fields[property] = value
		}
	}
	return fields
}

func getAvailableServiceIds() []string {
	pwd, _ := os.Getwd()
	servicesJson, err := gabs.ParseJSONFile(pwd + SERVICES_FILE_PATH)
	if err != nil {
		panic(err)
	}

	services := []string{}

	for serviceKey := range servicesJson.S("services").ChildrenMap() {
		services = append(services, serviceKey+"_config_V1")
	}

	return services
}

func getServiceSchema(path string) map[string]*gabs.Container {
	yamlFile, err := ioutil.ReadFile("oas.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(yamlFile, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	m2 := dyno.ConvertMapI2MapS(m)
	data, err := json.Marshal(m2)
	if err != nil {
		panic(err)
	}
	//pwd, _ := os.Getwd()
	//shemasJson, err := gabs.ParseJSONFile(pwd + SCHEMAS_FILE_PATH)
	shemasJson, err := gabs.ParseJSON(data)
	if err != nil {
		panic(err)
	}

	return shemasJson.Path(path).ChildrenMap()
}

func convert(m map[interface{}]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m {
		switch v2 := v.(type) {
		case map[interface{}]interface{}:
			res[fmt.Sprint(k)] = convert(v2)
		default:
			res[fmt.Sprint(k)] = v
		}
	}
	return res
}

func createFields(nodesMap map[string]*gabs.Container) map[string]*schema.Schema {
	fields := make(map[string]*schema.Schema)

	for key, node := range nodesMap {
		if _, ok := sensitiveFields[key]; ok {
			fields[key] = &schema.Schema{
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
		case INT_FIELD:
			nodeSchema.Type = schema.TypeInt
		case BOOL_FIELD:
			nodeSchema.Type = schema.TypeBool
		case ARRAY_FIELD:
			nodeSchema = getArrayFieldSchema(node)
		}
		fields[key] = nodeSchema
	}

	return fields
}

func getArrayFieldSchema(node *gabs.Container) *schema.Schema {
	itemType := node.Path("items.type").Data()

	childrenMap := node.Path("items.properties").ChildrenMap()

	arraySchema := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		}}

	if itemType == OBJECT_FIELD && len(childrenMap) > 0 {
		childrenSchemaMap := createFields(childrenMap)

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
