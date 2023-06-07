package fivetran

import (
	"os"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const SCHEMAS_PATH = "components.schemas."
const PROPERTIES_PATH = ".properties.config.properties"
const SERVICES_FILE_PATH = "/services.json"
const SCHEMAS_FILE_PATH = "/open-api-spec.json"

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
	"value":              true,
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

	schemaJson := getSchemaJson()

	fields := make(map[string]*schema.Schema)
	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH

		serviceSchema := schemaJson.Path(path).ChildrenMap()
		serviceFields := createFields(serviceSchema)
		for property, value := range serviceFields {
			if existingValue, ok := fields[property]; ok {
				if existingValue.Type != value.Type {
					property = service + property
				} else if existingValue.Type == schema.TypeSet {
					if _, ok := value.Elem.([]string); ok {
						if _, ok := existingValue.Elem.([]string); ok {

						} else {
							property = service + "." + property
						}
					}
					value, ok = updateExistingValue(existingValue, value)
					if !ok {
						property = strings.ToLower(service + "_" + property)
					}
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

func getSchemaJson() *gabs.Container {
	pwd, _ := os.Getwd()
	shemaJson, err := gabs.ParseJSONFile(pwd + SCHEMAS_FILE_PATH)
	if err != nil {
		panic(err)
	}

	return shemaJson
}

func createFields(nodesMap map[string]*gabs.Container) map[string]*schema.Schema {
	fields := make(map[string]*schema.Schema)

	for key, node := range nodesMap {
		if _, ok := sensitiveFields[key]; ok {
			fields[key] = &schema.Schema{
				Type:      schema.TypeString,
				Computed:  true,
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
		Type:     schema.TypeSet,
		Optional: true,
		Computed: true,
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

func updateExistingValue(existingValue *schema.Schema, newValue *schema.Schema) (*schema.Schema, bool) {
	if existingSchemaResourceValue, ok := existingValue.Elem.(*schema.Resource); ok {
		if newSchemaResourceValue, ok := newValue.Elem.(*schema.Resource); ok {
			for newSchemaResourceKey, newSchemaResourceValue := range newSchemaResourceValue.Schema {
				existingSchemaResourceValue.Schema[newSchemaResourceKey] = newSchemaResourceValue
			}
			existingValue.Elem = existingSchemaResourceValue
			return existingValue, true
		}
	}
	return existingValue, false
}
