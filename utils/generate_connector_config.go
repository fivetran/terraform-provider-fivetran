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
const PROPERTIES_PATH = "_config_V1.properties.config.properties"
const AUTH_PROPERTIES_PATH = "_config_V1.properties.auth.properties"
const SERVICES_PATH = "components.schemas.NewConnectorRequestV1.discriminator.mapping"

func main() {
	fmt.Println("Reading OAS...")
	schemaContainer := getSchemaJson()
	services := updateServices(schemaContainer)

	fmt.Println("Updating config fields")

	updateFields(services, schemaContainer,
		"common/fields.json",
		PROPERTIES_PATH,
		"common/fields-updated.json",
		"config-changes.txt",
	)

	fmt.Println("Updating auth fields")

	updateFields(services, schemaContainer,
		"common/auth-fields.json",
		AUTH_PROPERTIES_PATH,
		"common/auth-fields-updated.json",
		"auth-changes.txt",
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
) {
	fmt.Println("Reading existing " + existingFieldsFile + " file...")
	fieldsExisting := loadExistingFields(existingFieldsFile)
	fmt.Println("Loading updated fields")

	updated, changedFields := importFields(services, schemaContainer, fieldsExisting, schemaPropsPath)

	if updated {
		fmt.Println("New fields detected...")
		writeChangelog(changedFields, changelogFile)
		writeFields(fieldsExisting, updatedFieldsFile)
	} else {
		fmt.Println("No changes detected.")
	}
}

func loadExistingFields(file string) map[string]common.ConfigField {
	fmt.Println("Reading existing fields.json file...")
	content, err := os.ReadFile(file)

	fieldsExisting := make(map[string]common.ConfigField)

	if err != nil {
		fmt.Println("Failed to read file. Will create a new file by OAS.")
	} else {
		fmt.Println("Loading existing fields...")
		err = json.Unmarshal(content, &fieldsExisting)
		if err != nil {
			fmt.Println("Reading existing fields... Failed! File `fields.json` has wrong format.")
			panic(err)
		}
		fmt.Println("Reading existing fields... Success")
	}
	return fieldsExisting
}

func writeChangelog(changedFields map[string]common.ConfigField, clFile string) {
	var changeLog []string

	for fn, f := range changedFields {
		if fn != "schema" && fn != "table" && fn != "schema_prefix" {
			services := make([]string, 0, len(f.Description))
			for k := range f.Description {
				services = append(services, "`"+k+"`")
			}
			changeLog = append(changeLog, fmt.Sprintf("- Added field `fivetran_connector.config.%s` for services: %s.", fn, strings.Join(services, ", ")))
		}
	}
	//"config-changes.txt"
	err := os.WriteFile(clFile, []byte(strings.Join(changeLog, "\n")), 0644)
	if err != nil {
		fmt.Println("Failed to save changelog...")
		log.Fatal(err)
	}
}

func writeFields(fieldsExisting map[string]common.ConfigField, fileName string) {
	jsonResult, err := json.MarshalIndent(fieldsExisting, "", "   ")
	if err != nil {
		fmt.Println(err)
	}
	//"common/fields-updated.json"
	err = os.WriteFile(fileName, jsonResult, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated " + fileName)
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

func updateServices(schemaContainer *gabs.Container) []string {
	servicesOld, err := readLines("services.txt")

	if err != nil {
		fmt.Println("Failed to read existing services...")
		log.Fatal(err)
	}

	services := getAvailableServiceIds(schemaContainer)

	newServices := make([]string, 0)

	for _, s := range services {
		if _, ok := servicesOld[s]; !ok {
			newServices = append(newServices, fmt.Sprintf("- Supported service: `%s`", s))
		}
	}

	err = os.WriteFile("services-changelog.txt", []byte(strings.Join(newServices, "\n")), 0644)
	if err != nil {
		fmt.Println("Failed to save services changelog...")
		log.Fatal(err)
	}

	err = os.WriteFile("services-new.txt", []byte(strings.Join(services, "\n")), 0644)
	if err != nil {
		fmt.Println("Failed to save services list...")
		log.Fatal(err)
	}

	return services
}

func importFields(
	services []string,
	schemaContainer *gabs.Container,
	existingFields map[string]common.ConfigField,
	propPath string) (bool, map[string]common.ConfigField) {
	updated := false
	changeLog := make(map[string]common.ConfigField)

	for _, service := range services {
		path := SCHEMAS_PATH + service + propPath
		serviceSchema := schemaContainer.Path(path).ChildrenMap()
		serviceFieldsMap := createFields(serviceSchema, service)

		for name, field := range serviceFieldsMap {
			fmt.Println("INFO: processing field " + name + " (service " + service + ")")
			if existingField, ok := existingFields[name]; ok {
				fmt.Println("INFO: conflict detected - field " + name + " already exists in resouce schema.")
				if ableToMergeFields(field, existingField) {
					fmt.Println("INFO: field " + name + " will be merged.")
					m, u := mergeFields(field, existingField, service, changeLog, name)
					if fieldCouldBeIncluded(m) {
						existingFields[name] = m
						updated = updated || u
					} else {
						fmt.Println("INFO: field " + name + " ignored as inconsistent. Service: " + service + ".")
					}
				} else {
					serviceSpecificName := name + "_" + service
					fmt.Println("INFO: field " + name + " can't be merged with existing field. It will be mapped into " + serviceSpecificName + ".")
					if existingServiceField, ok := existingFields[serviceSpecificName]; ok {
						if ableToMergeFields(existingServiceField, field) {
							fmt.Println("INFO: Found existing service-specific field " + serviceSpecificName +
								". Field " + name + " will be merged with service-specific field.")
							mergedField, u := mergeFields(field, existingServiceField, service, changeLog, serviceSpecificName)
							mergedField.ApiField = name
							if fieldCouldBeIncluded(mergedField) {
								existingFields[serviceSpecificName] = mergedField
								updated = updated || u
							} else {
								fmt.Println("INFO: field " + name + " ignored as inconsistent. Service: " + service + ".")
							}
						} else {
							panic("ERROR: Unable to handle field " + name + " for " + service)
						}
					} else {
						if fieldCouldBeIncluded(field) {
							fmt.Println("INFO: field " + name + " will be added into schema as " + serviceSpecificName + ".")
							field.ApiField = name
							existingFields[serviceSpecificName] = field
							changeLog[serviceSpecificName] = field
							updated = true
						} else {
							fmt.Println("INFO: field " + name + " ignored as inconsistent. Service: " + service + ".")
						}
					}
				}
			} else {
				if fieldCouldBeIncluded(field) {
					fmt.Println("INFO: field " + name + " will be added into schema as " + name + ".")
					changeLog[name] = field
					existingFields[name] = field
					updated = true
				} else {
					fmt.Println("INFO: field " + name + " ignored as inconsistent. Service: " + service + ".")
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
			if ed != nd {
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
		updatedDescriptoion := appendFieldDescription(&newField, &existingField, service)
		updatedItemType := appendItemType(&newField, &existingField, service)
		return existingField, updatedDescriptoion || updatedItemType
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
			fieldInfo = getObjectField(node.Path("properties").ChildrenMap(), service, node.Path("enum").Data())
			fieldInfo.FieldValueType = common.Object
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
		return getObjectField(node.Path("items.properties").ChildrenMap(), service, node.Path("items.enum").Data())
	} else if itemType == INT_FIELD {
		field.FieldValueType = common.StringList
		field.ItemType[service] = common.Integer
	}

	return field
}

func getObjectField(childrenMap map[string]*gabs.Container, service string, enumElements interface{}) common.ConfigField {
	field := common.NewconfigField()
	if len(childrenMap) > 0 {
		field.FieldValueType = common.ObjectList

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
			if len(possibleItemKeys) == 0 {
				fmt.Println("WARNING: No key fields detected! Drifting changes possible.")
			} else if len(possibleItemKeys) > 1 {
				fmt.Println("WARNING: Multiple key fields detected! Please choose one manually.")
				field.ItemKeyField = "[" + strings.Join(possibleItemKeys, ", ") + "]"
			} else {
				field.ItemKeyField = possibleItemKeys[0]
			}
		}
	} else {
		if enumElements != nil {
			fmt.Println("ENUM-object: Object field without sub-fields but with enum.")
			field.FieldValueType = common.StringList
			field.Nullable = false
		} else {
			fmt.Println("WARNING: Object field without sub-fields.")
			field.FieldValueType = common.ObjectList
		}
	}
	return field
}

func getSchemaJson() *gabs.Container {
	fmt.Println("Reading OAS file...")
	oasJson, err := os.ReadFile("open-api-spec.json")

	if err != nil {
		fmt.Println("Reading OAS file... Failed to read schema file")
		panic(err)
	}

	shemaJson, err := gabs.ParseJSON(oasJson)

	if err != nil {
		fmt.Println("Reading OAS file... Failed to parse json schema")
		panic(err)
	}

	fmt.Println("Reading OAS file... Success")
	return shemaJson
}

func getAvailableServiceIds(schemaContainer *gabs.Container) []string {
	services := []string{}

	fileMap := schemaContainer.Path(SERVICES_PATH).ChildrenMap()

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
