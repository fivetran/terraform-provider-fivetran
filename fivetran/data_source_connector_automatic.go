package fivetran

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/Jeffail/gabs/v2"
	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const SCHEMAS_PATH = "schemas."
const PROPERTIES_PATH = ".properties.config.properties"

func dataSourceConnectorAutomatic() *schema.Resource {
	var result = &schema.Resource{
		ReadContext: dataSourceConnectorAutomaticRead,
		Schema: map[string]*schema.Schema{
			"id":                 {Type: schema.TypeString, Required: true},
			"group_id":           {Type: schema.TypeString, Computed: true},
			"service":            {Type: schema.TypeString, Computed: true},
			"service_version":    {Type: schema.TypeString, Computed: true},
			"name":               {Type: schema.TypeString, Computed: true},
			"destination_schema": dataSourceConnectorAutomaticDestinationSchemaSchema(),
			"connected_by":       {Type: schema.TypeString, Computed: true},
			"created_at":         {Type: schema.TypeString, Computed: true},
			"succeeded_at":       {Type: schema.TypeString, Computed: true},
			"failed_at":          {Type: schema.TypeString, Computed: true},
			"sync_frequency":     {Type: schema.TypeString, Computed: true},
			"daily_sync_time":    {Type: schema.TypeString, Computed: true},
			"schedule_type":      {Type: schema.TypeString, Computed: true},
			"paused":             {Type: schema.TypeString, Computed: true},
			"pause_after_trial":  {Type: schema.TypeString, Computed: true},
			"status":             dataSourceConnectorAutomaticSchemaStatus(),
			"config":             dataSourceConnectorAutomaticSchemaConfig(),
		},
	}
	return result
}

func dataSourceConnectorAutomaticDestinationSchemaSchema() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name":   {Type: schema.TypeString, Computed: true},
				"table":  {Type: schema.TypeString, Computed: true},
				"prefix": {Type: schema.TypeString, Computed: true},
			},
		},
	}
}

func dataSourceConnectorAutomaticSchemaStatus() *schema.Schema {
	return &schema.Schema{Type: schema.TypeList, Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"setup_state":        {Type: schema.TypeString, Computed: true},
				"sync_state":         {Type: schema.TypeString, Computed: true},
				"update_state":       {Type: schema.TypeString, Computed: true},
				"is_historical_sync": {Type: schema.TypeString, Computed: true},
				"tasks": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"code":    {Type: schema.TypeString, Computed: true},
							"message": {Type: schema.TypeString, Computed: true},
						},
					},
				},
				"warnings": {Type: schema.TypeList, Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"code":    {Type: schema.TypeString, Computed: true},
							"message": {Type: schema.TypeString, Computed: true},
						},
					},
				},
			},
		},
	}
}

func dataSourceConnectorAutomaticSchemaConfig() *schema.Schema {
	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		oasProperties := getDataSourceProperties(path)
		for key, value := range oasProperties {
			if existingValue, ok := properties[key]; ok {
				// if k == "reports" {
				// 	fmt.Printf("Type of val.Elem is %T\n", val.Elem)
				// }
				if existingValue.Type == schema.TypeList {
					if existingSchemaResourceValue, ok := existingValue.Elem.(*schema.Resource); ok {
						if newSchemaResourceValue, ok := value.Elem.(*schema.Resource); ok {
							for kY, vY := range newSchemaResourceValue.Schema {
								existingSchemaResourceValue.Schema[kY] = vY
							}
							existingValue.Elem = existingSchemaResourceValue
							properties[key] = existingValue
							continue
						}
					} else if _, ok := existingValue.Elem.(*schema.Schema); ok {
						// if v3, ok := schemaSchemaValue.Elem.(*schema.Resource); ok {
						// 	if vX1, ok := value.Elem.(*schema.Resource); ok {
						// 		for kY, vY := range vX1.Schema {
						// 			v3.Schema[kY] = vY
						// 		}
						// 		existingValue.Elem = v3
						// 		properties[key] = existingValue
						// 		continue
						// 	}
						// }
					} else if _, ok := existingValue.Elem.(map[string]*schema.Schema); ok {
						continue
					}
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

func getDataSourceProperties(path string) map[string]*schema.Schema {
	shemasJson, err := gabs.ParseJSONFile("/Users/lukadevic/Fivetran/terraform-provider-fivetran/fivetran/schemas.json")
	if err != nil {
		panic(err)
	}

	properties := make(map[string]*schema.Schema)

	for key, child := range shemasJson.Path(path).ChildrenMap() {
		value := &schema.Schema{
			Type:     schema.TypeString,
			Computed: true}

		propertyType := child.Search("type").Data()

		switch propertyType {
		case "object":
			value = &schema.Schema{
				Type:     schema.TypeString,
				Computed: true}
		case "integer":
			value = &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true}
		case "boolean":
			value = &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true}
		case "array":
			itemType := child.Path("items.type").Data()

			childrenMap := child.Path("items.properties").ChildrenMap()

			if itemType == "object" && len(childrenMap) > 0 {
				childrenSchemaMap := make(map[string]*schema.Schema)

				for key2, child2 := range childrenMap {
					value2 := &schema.Schema{
						Type:     schema.TypeString,
						Computed: true}
					propertyType2 := child2.Search("type").Data()
					switch propertyType2 {
					case "object":
						value2 = &schema.Schema{
							Type:     schema.TypeString,
							Computed: true}
					case "integer":
						value2 = &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true}
					case "boolean":
						value2 = &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true}
					case "array":
						itemType2 := child2.Path("items.type").Data()

						if itemType2 == "string" || itemType2 == "object" {
							value2 = &schema.Schema{
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								}}
						}

						if itemType2 == "integer" {
							value2 = &schema.Schema{
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								}}
						}
					default:
						value2 = &schema.Schema{
							Type:     schema.TypeString,
							Computed: true}
					}

					childrenSchemaMap[key2] = value2
				}

				value = &schema.Schema{
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: childrenSchemaMap,
					},
				}
			} else if itemType == "string" || itemType == "object" {
				value = &schema.Schema{
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					}}
			}

			if itemType == "integer" {
				value = &schema.Schema{
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					}}
			}
		default:
			value = &schema.Schema{
				Type:     schema.TypeString,
				Computed: true}

		}
		properties[key] = value
	}

	return properties
}

func dataSourceConnectorAutomaticRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*fivetran.Client)

	resp, err := client.NewConnectorDetails().ConnectorID(d.Get("id").(string)).DoCustomMerged(ctx)
	if err != nil {
		return newDiagAppend(diags, diag.Error, "service error", fmt.Sprintf("%v; code: %v; message: %v", err, resp.Code, resp.Message))
	}

	// msi stands for Map String Interface
	msi := make(map[string]interface{})
	mapAddStr(msi, "id", resp.Data.ID)
	mapAddStr(msi, "group_id", resp.Data.GroupID)
	mapAddStr(msi, "service", resp.Data.Service)
	mapAddStr(msi, "service_version", intPointerToStr(resp.Data.ServiceVersion))
	mapAddStr(msi, "name", resp.Data.Schema)
	mapAddXInterface(msi, "destination_schema", dataSourceConnectorAutomaticReadDestinationSchema(resp.Data.Schema, resp.Data.Service))
	mapAddStr(msi, "connected_by", resp.Data.ConnectedBy)
	mapAddStr(msi, "created_at", resp.Data.CreatedAt.String())
	mapAddStr(msi, "succeeded_at", resp.Data.SucceededAt.String())
	mapAddStr(msi, "failed_at", resp.Data.FailedAt.String())
	mapAddStr(msi, "sync_frequency", intPointerToStr(resp.Data.SyncFrequency))
	mapAddStr(msi, "daily_sync_time", resp.Data.DailySyncTime)
	mapAddStr(msi, "schedule_type", resp.Data.ScheduleType)
	mapAddStr(msi, "paused", boolPointerToStr(resp.Data.Paused))
	mapAddStr(msi, "pause_after_trial", boolPointerToStr(resp.Data.PauseAfterTrial))
	mapAddXInterface(msi, "status", dataSourceConnectorAutomaticReadStatus(&resp))
	mapAddXInterface(msi, "config", dataSourceConnectorAutomaticReadConfig(&resp))
	for k, v := range msi {
		if err := d.Set(k, v); err != nil {
			return newDiagAppend(diags, diag.Error, "set error", fmt.Sprint(err))
		}
	}

	d.SetId(resp.Data.ID)

	return diags
}

func dataSourceConnectorAutomaticReadDestinationSchema(schema string, service string) []interface{} {
	return readDestinationSchema(schema, service)
}

// dataSourceConnectorReadStatus receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "status" list.
func dataSourceConnectorAutomaticReadStatus(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	status := make([]interface{}, 1)

	s := make(map[string]interface{})
	mapAddStr(s, "setup_state", resp.Data.Status.SetupState)
	mapAddStr(s, "sync_state", resp.Data.Status.SyncState)
	mapAddStr(s, "update_state", resp.Data.Status.UpdateState)
	mapAddStr(s, "is_historical_sync", boolPointerToStr(resp.Data.Status.IsHistoricalSync))
	mapAddXInterface(s, "tasks", dataSourceConnectorReadStatusFlattenTasks(resp))
	mapAddXInterface(s, "warnings", dataSourceConnectorReadStatusFlattenWarnings(resp))
	status[0] = s

	return status
}

// dataSourceConnectorReadConfig receives a *fivetran.ConnectorDetailsResponse and returns a []interface{}
// containing the data type accepted by the "config" list.
func dataSourceConnectorAutomaticReadConfig(resp *fivetran.ConnectorCustomMergedDetailsResponse) []interface{} {
	config := make([]interface{}, 1)

	configMap := make(map[string]interface{})

	c := resp.Data.CustomConfig

	m := structToMap(resp.Data.Config)
	for key, value := range m {
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice && reflect.TypeOf(value).Elem().Kind() != reflect.String {
			fmt.Printf("Type of value is %T\n", value)

			var out []interface{}
			for i := 0; i < rv.Len(); i++ {
				out = append(out, rv.Index(i).Interface())
			}

			adb := structToMap(out[0])
			log.Output(1, intToStr(len(adb)))
			out[0] = adb
			c[key] = out
			continue
		}

		c[key] = value
	}

	services := getAvailableServiceIds()

	properties := make(map[string]*schema.Schema)

	for _, service := range services {
		path := SCHEMAS_PATH + service + PROPERTIES_PATH
		newProperties := getDataSourceProperties(path)
		for k, v := range newProperties {
			properties[k] = v
		}
	}

	for key, value := range properties {

		// if key == "adobe_analytics_configurations" {
		// 	fmt.Printf("Type of c[key] is %T\n", c[key])
		// }
		if value.Type == schema.TypeSet || value.Type == schema.TypeList {
			if v, ok := c[key].([]string); ok {
				configMap[key] = xStrXInterface(v)
				continue
			}
			if v, ok := c[key].([]interface{}); ok {
				if v2, ok := v[0].(map[string]interface{}); ok {
					log.Output(2, intToStr(len(v2)))
					configMap[key] = v
				} else {
					configMap[key] = xInterfaceStrXStr(v)
				}
				continue
			}
		}
		if v, ok := c[key].(string); ok && v != "" {
			valueType := value.Type
			switch valueType {
			case schema.TypeBool:
				configMap[key] = strToBool(v)
			case schema.TypeInt:
				configMap[key] = strToInt(v)
			default:
				configMap[key] = v
			}
		}
	}
	config[0] = configMap

	return config
}

/*
This function will convert object from struct to map[string]interface{}
based on JSON tag in structs.
*/
func structToMap(item interface{}) map[string]interface{} {

	res := map[string]interface{}{}
	if item == nil {
		return res
	}
	v := reflect.TypeOf(item)
	reflectValue := reflect.ValueOf(item)
	reflectValue = reflect.Indirect(reflectValue)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("json")
		field := reflectValue.Field(i).Interface()
		if tag != "" && tag != "-" {
			if v.Field(i).Type.Kind() == reflect.Struct {
				res[tag] = structToMap(field)
			} else {
				res[tag] = field
			}
		}
	}
	return res
}