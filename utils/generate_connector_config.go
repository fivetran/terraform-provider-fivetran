package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/fivetran/terraform-provider-fivetran/fivetran/common"
)

const OBJECT_FIELD = "object"
const INT_FIELD = "integer"
const BOOL_FIELD = "boolean"
const ARRAY_FIELD = "array"
const STRING_FIELD = "string"

const SCHEMAS_PATH = "components.schemas."

func main() {
	fmt.Println("Reading OAS...")
	schemaContainer := getSchemaJson()

	services := updateServices(
		schemaContainer,
		"components.schemas.NewConnectorRequestV1.discriminator.mapping",
		"services.txt",
		"services-changelog.txt",
		"services-new.txt",
		"New connection services supported:",
	)


	fmt.Println("Updating config fields")

	updateFields(services, schemaContainer,
		"fivetran/common/fields.json",
		"_config_V1.properties.config.properties",
		"fivetran/common/fields-updated.json",
		"config-changes.txt",
		false,
		"New connection config fields supported:",
	)
	fmt.Println("Updating schema fields")

	updateFields(services, schemaContainer,
		"fivetran/common/fields-updated.json",
		"schema_format_schema_table.properties.config.properties",
		"fivetran/common/fields-updated.json",
		"config-changes-schema_format_schema_table.txt",
		false,
		"New connection config fields supported:",
	)

	updateFields(services, schemaContainer,
		"fivetran/common/fields-updated.json",
		"schema_format_schema_prefix.properties.config.properties",
		"fivetran/common/fields-updated.json",
		"config-changes-schema_format_schema_prefix.txt",
		false,
		"New connection config fields supported:",
	)

	updateFields(services, schemaContainer,
		"fivetran/common/fields-updated.json",
		"schema_format_schema_table_group.properties.config.properties",
		"fivetran/common/fields-updated.json",
		"config-changes-schema_format_schema_table_group.txt",
		false,
		"New connection config fields supported:",
	)

	updateFields(services, schemaContainer,
		"fivetran/common/fields-updated.json",
		"schema_format_schema.properties.config.properties",
		"fivetran/common/fields-updated.json",
		"config-changes-schema_format_schema.txt",
		false,
		"New connection config fields supported:",
	)

	fmt.Println("Updating auth fields")

	updateFields(services, schemaContainer,
		"fivetran/common/auth-fields.json",
		"_config_V1.properties.auth.properties",
		"fivetran/common/auth-fields-updated.json",
		"auth-changes.txt",
		true,
		"New connection auth fields supported:",
	)

	fmt.Println("Updating Destinations config fields")

	destinationServices := updateServices(
		schemaContainer,
		"components.schemas.NewDestinationRequest.discriminator.mapping",
		"destination-services.txt",
		"destination-services-changelog.txt",
		"destination-services-new.txt",
		"New destination services supported:",
	)

	updateFields(destinationServices, schemaContainer,
		"fivetran/common/destination-fields.json",
		"_config_V1.properties.config.properties",
		"fivetran/common/destination-fields-updated.json",
		"destination-config-changes.txt",
		true,
		"New destination config fields supported:",
	)

	fmt.Println("Done!")
}

func updateFields(
	services []string,
	schemaContainer *gabs.Container,
	existingFieldsFile string,
	schemaPropsPath string,
	updatedFieldsFile string,
	changelogFile string,
	isDestination bool,
	title string,
) {
	fieldsExisting := loadExistingFields(existingFieldsFile)

	updated, changedFields := importFields(services, schemaContainer, fieldsExisting, schemaPropsPath, isDestination)

	if updated {
		writeChangelog(changedFields, changelogFile, isDestination, title)
		writeFields(fieldsExisting, updatedFieldsFile)
	}
}

func loadExistingFields(file string) map[string]common.ConfigField {
	content, err := os.ReadFile(file)
	fieldsExisting := make(map[string]common.ConfigField)

	if err != nil {
		log.Fatal(err)
	} else {
		err = json.Unmarshal(content, &fieldsExisting)
		if err != nil {
			log.Fatal(err)
		}
	}
	return fieldsExisting
}

func writeChangelog(changedFields map[string]common.ConfigField, clFile string, isDestination bool, title string) {
	var changeLog []string
	changeLog = append(changeLog, title)

	var resourceType string
	if isDestination {
		resourceType = "fivetran_destination"
	} else {
		resourceType =  "fivetran_connector"
	}

	for fn, f := range changedFields {
		if fn != "schema" && fn != "table" && fn != "schema_prefix" && fn != "table_group_name" {
			services := make([]string, 0, len(f.Description))
			for k := range f.Description {
				services = append(services, "`"+k+"`")
			}
			changeLog = append(changeLog, fmt.Sprintf("- Added field `%s.config.%s` for services: %s.", resourceType, fn, strings.Join(services, ", ")))
		} else {
			services := make([]string, 0, len(f.Description))
			for k := range f.Description {
				services = append(services, "`"+k+"`")
			}
			changeLog = append(changeLog, fmt.Sprintf("- Added field `%s.destination_schema.%s` for services: %s.", resourceType, fn, strings.Join(services, ", ")))
		}
	}

	err := os.WriteFile(clFile, []byte(strings.Join(changeLog, "\n")), 0644); 
	if err != nil {
    	log.Fatal(err)
	}
}

func writeFields(fieldsExisting map[string]common.ConfigField, fileName string) {
	jsonResult, err := json.MarshalIndent(fieldsExisting, "", "   ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(fileName, jsonResult, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func readLines(fileName string) (map[string]bool, error) {
	result := make(map[string]bool)
	file, err := os.Open(fileName)
	if err != nil {
		return result, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result[scanner.Text()] = true
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}
	return result, nil
}

func updateServices(schemaContainer *gabs.Container, servicesPath, servicesFile, changelogFile, newServicesFile string, title string) []string {
	servicesOld, err := readLines(servicesFile)

	if err != nil {
		log.Fatal(err)
	}

	services := getAvailableServiceIds(schemaContainer, servicesPath)

	newServices := make([]string, 0)
	newServices = append(newServices, title)
	for _, s := range services {
		if _, ok := servicesOld[s]; !ok {
			newServices = append(newServices, fmt.Sprintf("- Supported service: `%s`", s))
		}
	}

	err = os.WriteFile(changelogFile, []byte(strings.Join(newServices, "\n")), 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(newServicesFile, []byte(strings.Join(services, "\n")), 0644)
	if err != nil {
		log.Fatal(err)
	}

	return services
}

func checkSchemaAlignment(schemaContainer *gabs.Container, service string, path string) bool {
	fileMap := schemaContainer.Path("components.schemas." + service + "_NewConnectorRequestV1").ChildrenMap()
	for _, f := range fileMap {
		for _, f2 := range f.Children() {
			prefix, _, _ := strings.Cut(path, ".")
			path, _ := strings.CutSuffix(f2.Path("$ref").String(), "\"")
			if strings.HasSuffix(path, prefix) {
				return true
			}
		}
	}

	return false
}

func importFields(
	services []string,
	schemaContainer *gabs.Container,
	existingFields map[string]common.ConfigField,
	propPath string,
	isDestination bool) (bool, map[string]common.ConfigField) {
	updated := false
	changeLog := make(map[string]common.ConfigField)

	for _, service := range services {
		var path string
		if strings.HasPrefix(propPath, "_") {
			path = service + propPath
		} else {
			path = propPath
		}

		if !isDestination && !checkSchemaAlignment(schemaContainer, service, path) {
			continue
		}

		path = SCHEMAS_PATH + path

		serviceSchema := schemaContainer.Path(path).ChildrenMap()
		serviceFieldsMap := createFields(serviceSchema, service)

		for name, field := range serviceFieldsMap {
			if existingField, ok := existingFields[name]; ok {
				if ableToMergeFields(field, existingField) {
					m, u := mergeFields(field, existingField, service, changeLog, name)

					if fieldCouldBeIncluded(m) {
						existingFields[name] = m
						updated = updated || u
					}
				} else {
					serviceSpecificName := name + "_" + service
					if existingServiceField, ok := existingFields[serviceSpecificName]; ok {
						if ableToMergeFields(existingServiceField, field) {
							mergedField, u := mergeFields(field, existingServiceField, service, changeLog, serviceSpecificName)
							mergedField.ApiField = name
							if fieldCouldBeIncluded(mergedField) {
								existingFields[serviceSpecificName] = mergedField
								updated = updated || u
							}
						} else {
							log.Fatal("ERROR: Unable to handle field " + name + " for " + service)
						}
					} else {
						if fieldCouldBeIncluded(field) {
							field.ApiField = name
							existingFields[serviceSpecificName] = field
							changeLog[serviceSpecificName] = field
							updated = true
						}
					}
				}
			} else {
				if fieldCouldBeIncluded(field) {
					changeLog[name] = field
					existingFields[name] = field
					updated = true
				}
			}
		}
	}
	return updated, changeLog
}

func fieldCouldBeIncluded(field common.ConfigField) bool {
	// Do not include empty object-fields
	if field.FieldValueType == common.ObjectList && len(field.ItemFields) == 0 {
		return false
	}

	return true
}

func appendFieldDescription(newField, existingField *common.ConfigField, service string) bool {
	// check if existing field already has description for this service
	if ed, ok := existingField.Description[service]; ok {
		// check if new field has description
		if nd, ok := newField.Description[service]; ok {
			// check if description updated in new field
			if ed != nd && nd != "" {
				existingField.Description[service] = nd
				return true
			}
		}
		// check if new field has description
	} else if nd, ok := newField.Description[service]; ok {
		existingField.Description[service] = nd
		return true
	}
	return false
}

func appendItemType(newField, existingField *common.ConfigField, service string) bool {
	if existingField.ItemType == nil {
		existingField.ItemType = make(map[string]common.FieldValueType)
	}

	if ed, ok := existingField.ItemType[service]; ok {
		if nd, ok := newField.ItemType[service]; ok {
			if ed != nd {
				existingField.ItemType[service] = nd
				return true
			}
		}
	} else if nd, ok := newField.ItemType[service]; ok {
		existingField.ItemType[service] = nd
		return true
	}
	return false
}

func appendSensitiveExclusion(newField, existingField *common.ConfigField, service string) bool {
	if newField.Sensitive == existingField.Sensitive || !newField.Sensitive {
		return false
	}
	if existingField.SensitiveExclusions == nil {
		existingField.SensitiveExclusions = map[string]bool{}
	}
	existingField.SensitiveExclusions[service] = newField.Sensitive
	return true
}

func ableToMergeFields(a, b common.ConfigField) bool {
	if a.FieldValueType != b.FieldValueType {
		// can't merge fields of different types
		return false
	}
	if a.FieldValueType != common.ObjectList {
		// for primitive types we can assume this case as "mergable"
		return true
	}

	// we should check if we are able to merge sub-fields for object collections
	for ak, av := range a.ItemFields {
		// try to get corresponding fields between a and b
		if bv, ok := b.ItemFields[ak]; ok {
			// if the field has the same key in both models - check if we able to merge them
			if !ableToMergeFields(av, bv) {
				return false
			}
		}
	}

	return true
}

func mergeFields(newField, existingField common.ConfigField, service string, changeLog map[string]common.ConfigField, parentName string) (common.ConfigField, bool) {
	if existingField.FieldValueType != common.ObjectList && existingField.FieldValueType != common.Object {
		// there's nothing to merge for primitive types, there won't be any updates in schema
		updatedDescription := appendFieldDescription(&newField, &existingField, service)
		updatedItemType := appendItemType(&newField, &existingField, service)
		updatedSensitive := appendSensitiveExclusion(&newField, &existingField, service)

		return existingField, updatedDescription || updatedItemType || updatedSensitive
	}

	updated := false
	// we should just add all missing fields from a to b
	for nk, nv := range newField.ItemFields {
		ev, ok := existingField.ItemFields[nk]
		if ok {
			// we should merge existing fields (in case if it has sub-items with new fields)
			merged, upd := mergeFields(nv, ev, service, changeLog, parentName+"."+nk)
			existingField.ItemFields[nk] = merged
			if upd {
				updated = true
			}
		} else {
			// if field represented in a, but not in b schema will be updated
			existingField.ItemFields[nk] = nv
			changeLog[parentName+"."+nk] = nv
			updated = true
		}
	}

	updatedDescription := appendFieldDescription(&newField, &existingField, service)
	updatedItemType := appendItemType(&newField, &existingField, service)
	return existingField, updated || updatedDescription || updatedItemType
}

func createFields(nodesMap map[string]*gabs.Container, service string) map[string]common.ConfigField {
	fields := make(map[string]common.ConfigField)

	for key, node := range nodesMap {
		fieldInfo := common.NewconfigField()
		nodeType := node.Search("type").Data()

		switch nodeType {
		case STRING_FIELD:
			enumElements := node.Search("enum").Data()
			if enumElements != nil {
				fieldInfo.Nullable = false
			}
		case INT_FIELD:
			fieldInfo.FieldValueType = common.Integer
			fieldInfo.Nullable = false
		case BOOL_FIELD:
			fieldInfo.FieldValueType = common.Boolean
			fieldInfo.Nullable = false
		case ARRAY_FIELD:
			fieldInfo = getArrayFieldSchema(node, fieldInfo, service)
		case OBJECT_FIELD:
			fieldInfo = getObjectField(node.Path("properties").ChildrenMap(), service, node.Path("enum").Data(), false)
		}

		nodeDescription := node.Search("description").Data()

		if nodeDescription != nil {
			fieldInfo.Description[service] = processDescription(nodeDescription.(string))
		} else {
			fieldInfo.Description[service] = ""
		}

		nodeFormat := node.Search("format").Data()

		if nodeFormat != nil && nodeFormat == "password" {
			fieldInfo.Sensitive = true
		}

		fields[key] = fieldInfo
	}
	return fields
}

func getArrayFieldSchema(node *gabs.Container, field common.ConfigField, service string) common.ConfigField {
	itemType := node.Path("items.type").Data()
	if itemType == STRING_FIELD {
		field.FieldValueType = common.StringList
		field.ItemType[service] = common.String
	} else if itemType == OBJECT_FIELD {
		return getObjectField(node.Path("items.properties").ChildrenMap(), service, node.Path("items.enum").Data(), true)
	} else if itemType == INT_FIELD {
		field.FieldValueType = common.StringList
		field.ItemType[service] = common.Integer
	}

	return field
}

func getObjectField(childrenMap map[string]*gabs.Container, service string, enumElements interface{}, isArray bool) common.ConfigField {
	field := common.NewconfigField()
	if len(childrenMap) > 0 {
		if isArray {
			field.FieldValueType = common.ObjectList
		} else {
			field.FieldValueType = common.Object
		}

		needItemKey := false
		possibleItemKeys := make([]string, 0)

		field.ItemFields = make(map[string]common.ConfigField)

		for k, v := range createFields(childrenMap, service) {
			if v.Sensitive {
				needItemKey = true
			} else {
				possibleItemKeys = append(possibleItemKeys, k)
			}
			field.ItemFields[k] = v
		}

		if needItemKey {
			if len(possibleItemKeys) > 1 {
				field.ItemKeyField = "[" + strings.Join(possibleItemKeys, ", ") + "]"
			} else {
				field.ItemKeyField = possibleItemKeys[0]
			}
		}
	} else {
		if enumElements != nil {
			field.FieldValueType = common.String
			if isArray {
				field.FieldValueType = common.StringList
			}
			field.Nullable = false
		} else {
			field.FieldValueType = common.Object
			if isArray {
				field.FieldValueType = common.ObjectList
			}
		}
	}
	return field
}

func getSchemaJson() *gabs.Container {
	oasJson, err := os.ReadFile("open-api-spec.json")
	if err != nil {
		log.Fatal(err)
	}

	shemaJson, err := gabs.ParseJSON(oasJson)

	if err != nil {
		log.Fatal(err)
	}

	return shemaJson
}

func getAvailableServiceIds(schemaContainer *gabs.Container, servicesPath string) []string {
	services := []string{}

	fileMap := schemaContainer.Path(servicesPath).ChildrenMap()

	for serviceKey := range fileMap {
		services = append(services, serviceKey)
	}

	sort.Strings(services)

	return services
}

func processDescription(description string) string {
	description = strings.ReplaceAll(description, "](/docs/", "](https://fivetran.com/docs/")
	description = strings.ReplaceAll(description, "\u003cstrong\u003e", "")
	description = strings.ReplaceAll(description, "\u003c/strong\u003e", "")
	description = strings.ReplaceAll(description, "\u003cbr\u003e", "")
	description = strings.ReplaceAll(description, "â", "-")
	description = strings.ReplaceAll(description, "\u003c", "")
	description = strings.ReplaceAll(description, "\u003e", "")
	return description
}
